package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/url"
	"os"
	"path"
	"time"

	"github.com/bluesky-social/indigo/api/atproto"
	"github.com/bluesky-social/indigo/api/bsky"
	"github.com/bluesky-social/indigo/events"
	"github.com/bluesky-social/indigo/events/schedulers/sequential"
	"github.com/bluesky-social/indigo/lex/util"
	"github.com/bluesky-social/indigo/repo"
	"github.com/bluesky-social/indigo/repomgr"
	"github.com/gorilla/websocket"
	"github.com/pojntfx/atmosfeed/pkg/persisters"
)

func main() {
	pdsURL := flag.String("pds-url", "wss://bsky.social/", "PDS URL (can also be set using `PDS_URL` env variable)")
	postgresURL := flag.String("postgres-url", "postgresql://postgres@localhost:5432/atmosfeed?sslmode=disable", "PostgreSQL URL (can also be set using `POSTGRES_URL` env variable)")
	verbose := flag.Bool("verbose", false, "Whether to enable verbose logging")

	flag.Parse()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if v := os.Getenv("POSTGRES_URL"); v != "" {
		log.Println("Using database address from POSTGRES_URL env variable")

		*postgresURL = v
	}

	pu, err := url.Parse(*pdsURL)
	if err != nil {
		panic(err)
	}
	pu = pu.JoinPath("xrpc", "com.atproto.sync.subscribeRepos")

	atu, err := url.Parse("at://")
	if err != nil {
		panic(err)
	}

	conn, _, err := websocket.DefaultDialer.DialContext(ctx, pu.String(), nil)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	log.Println("Connected to PDS", *pdsURL)

	p := persisters.NewPersister(*postgresURL)

	if err := p.Init(); err != nil {
		panic(err)
	}

	log.Println("Connected to PostgreSQL")

	handlers := events.RepoStreamCallbacks{
		RepoCommit: func(c *atproto.SyncSubscribeRepos_Commit) error {
			rp, err := repo.ReadRepoFromCar(ctx, bytes.NewReader(c.Blocks))
			if err != nil {
				log.Println("Could not parse repo, skipping:", err)

				return nil
			}

		l:
			for _, op := range c.Ops {
				switch repomgr.EventKind(op.Action) {
				case repomgr.EvtKindCreateRecord:
					_, res, err := rp.GetRecord(ctx, op.Path)
					if err != nil {
						log.Println("Could not parse record, skipping:", err)

						continue l
					}

					d := util.LexiconTypeDecoder{
						Val: res,
					}

					b, err := d.MarshalJSON()
					if err != nil {
						log.Println("Could not marshal lexicon, skipping:", err)

						continue l
					}

					var post bsky.FeedPost
					if err := json.Unmarshal(b, &post); err != nil {
						log.Println("Could not unmarshal post, skipping:", err)

						continue l
					}

					if post.LexiconTypeID == "app.bsky.feed.post" {
						createdAt, err := time.Parse(time.RFC3339, post.CreatedAt)
						if err != nil {
							log.Println("Could not parse post date, skipping:", err)

							continue l
						}

						id, err := p.CreatePost(
							ctx,
							createdAt,
							rp.RepoDid(),
							path.Base(op.Path),
							post.Text,
							post.Reply != nil,
							post.Langs,
						)
						if err != nil {
							log.Println("Could not insert post, skipping:", err)

							continue l
						}

						if *verbose {
							log.Println(
								id,
								post.CreatedAt,
								atu.JoinPath(rp.RepoDid(), post.LexiconTypeID, path.Base(op.Path)),
								post.Text,
								post.Reply != nil,
								post.Langs,
								0,
							)
						}
					} else if post.LexiconTypeID == "app.bsky.feed.like" {
						var like bsky.FeedLike
						if err := json.Unmarshal(b, &like); err != nil {
							log.Println("Could not unmarshal like, skipping:", err)

							continue l
						}

						fmt.Println(
							"Liked:",
							post.CreatedAt,
							like.Subject.Uri,
						)
					}
				}
			}

			return nil
		},
	}

	if err := events.HandleRepoStream(
		ctx,
		conn,
		sequential.NewScheduler(
			conn.RemoteAddr().String(),
			handlers.EventHandler,
		),
	); err != nil {
		panic(err)
	}
}