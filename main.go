package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/url"
	"path"

	"github.com/bluesky-social/indigo/api/atproto"
	"github.com/bluesky-social/indigo/api/bsky"
	"github.com/bluesky-social/indigo/events"
	"github.com/bluesky-social/indigo/events/schedulers/sequential"
	"github.com/bluesky-social/indigo/lex/util"
	"github.com/bluesky-social/indigo/repo"
	"github.com/bluesky-social/indigo/repomgr"
	"github.com/gorilla/websocket"
)

func main() {
	raddr := flag.String("raddr", "wss://bsky.social/", "Remote address of the PDS to use")

	flag.Parse()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	pu, err := url.Parse(*raddr)
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

	log.Println("Connected to", *raddr)

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

					p := util.LexiconTypeDecoder{
						Val: res,
					}

					b, err := p.MarshalJSON()
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
						fmt.Println(
							"Created:",
							post.CreatedAt,
							atu.JoinPath(rp.RepoDid(), post.LexiconTypeID, path.Base(op.Path)),
							post.Text,
							post.Reply != nil,
							post.Langs,
						)
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
