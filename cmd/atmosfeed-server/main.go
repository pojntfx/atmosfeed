package main

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/url"
	"os"
	"path"
	"signature"
	"sync"
	"time"

	"github.com/bluesky-social/indigo/api/atproto"
	"github.com/bluesky-social/indigo/api/bsky"
	"github.com/bluesky-social/indigo/events"
	"github.com/bluesky-social/indigo/events/schedulers/sequential"
	lutil "github.com/bluesky-social/indigo/lex/util"
	"github.com/bluesky-social/indigo/repo"
	"github.com/bluesky-social/indigo/repomgr"
	iutil "github.com/bluesky-social/indigo/util"
	"github.com/gorilla/websocket"
	"github.com/lib/pq"
	"github.com/loopholelabs/scale"
	"github.com/loopholelabs/scale/scalefunc"
	"github.com/pojntfx/atmosfeed/pkg/models"
	"github.com/pojntfx/atmosfeed/pkg/persisters"
)

const (
	channelFeedInserted = "feed_inserted"
	channelFeedUpdated  = "feed_updated"
	channelFeedDeleted  = "feed_deleted"
)

func main() {
	pdsURL := flag.String("pds-url", "wss://bsky.social/", "PDS URL (can also be set using `PDS_URL` env variable)")
	postgresURL := flag.String("postgres-url", "postgresql://postgres@localhost:5432/atmosfeed?sslmode=disable", "PostgreSQL URL (can also be set using `POSTGRES_URL` env variable)")
	verbose := flag.Bool("verbose", false, "Whether to enable verbose logging")
	classifierTimeout := flag.Duration("classifier-timeout", time.Second, "Amount of time after which to stop a classifer Scale function from running")

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

	var classifierLock sync.Mutex
	classifiers := map[string]*scale.Instance[*signature.Signature]{}

	listener := pq.NewListener(*postgresURL, 10*time.Second, time.Minute, nil)
	if err := listener.Listen(channelFeedInserted); err != nil {
		panic(err)
	}

	if err := listener.Listen(channelFeedUpdated); err != nil {
		panic(err)
	}

	if err := listener.Listen(channelFeedDeleted); err != nil {
		panic(err)
	}

	go func() {
		for notification := range listener.Notify {
			switch notification.Channel {
			case channelFeedInserted:
				if *verbose {
					log.Println("Created feed", notification.Extra)
				}

				fallthrough

			case channelFeedUpdated:
				if *verbose {
					log.Println("Updated feed", notification.Extra)
				}

				func() {
					classifierSource, err := p.GetFeedClassifier(ctx, notification.Extra)
					if err != nil {
						log.Println("Could not fetch new classifier, skipping:", err)

						return
					}

					classifierLock.Lock()
					defer classifierLock.Unlock()

					fn := &scalefunc.Schema{}
					if err := fn.Decode(classifierSource); err != nil {
						panic(err)
					}

					runtime, err := scale.New(scale.NewConfig(signature.New).WithFunction(fn))
					if err != nil {
						panic(err)
					}

					instance, err := runtime.Instance()
					if err != nil {
						panic(err)
					}
					defer instance.Cleanup()

					classifiers[notification.Extra] = instance
				}()

			case channelFeedDeleted:
				if *verbose {
					log.Println("Deleted feed", notification.Extra)
				}

				func() {
					classifierLock.Lock()
					defer classifierLock.Unlock()

					delete(classifiers, notification.Extra)
				}()
			}
		}
	}()

	log.Println("Connected to PostgreSQL")

	classifierSources, err := p.GetFeeds(ctx)
	if err != nil {
		panic(err)
	}

	for _, classifierSource := range classifierSources {
		func() {
			classifierLock.Lock()
			defer classifierLock.Unlock()

			fn := &scalefunc.Schema{}
			if err := fn.Decode(classifierSource.Classifier); err != nil {
				panic(err)
			}

			runtime, err := scale.New(scale.NewConfig(signature.New).WithFunction(fn))
			if err != nil {
				panic(err)
			}

			instance, err := runtime.Instance()
			if err != nil {
				panic(err)
			}
			defer instance.Cleanup()

			classifiers[classifierSource.Name] = instance
		}()
	}

	log.Println("Fetched classifiers")

	classify := func(post models.Post) error {

		errs := make(chan error)

		classifierLock.Lock()
		defer classifierLock.Unlock()

		var wg sync.WaitGroup
		for feed, classifier := range classifiers {
			wg.Add(1)

			go func(feed string, classifier *scale.Instance[*signature.Signature]) {
				defer wg.Done()

				p := signature.NewPost()

				p.Did = post.Did
				p.Rkey = post.Rkey
				p.Text = post.Text

				p.Langs = post.Langs

				p.CreatedAt = post.CreatedAt.Unix()
				p.Likes = int64(post.Likes)

				p.Reply = post.Reply

				s := signature.New()
				s.Context.Post = p

				ctx, cancel := context.WithTimeout(context.Background(), *classifierTimeout)
				defer cancel()

				if err := classifier.Run(ctx, s); err != nil {
					errs <- err

					return
				}

				if s.Context.Include {
					fmt.Println(feed, post)
				}
			}(feed, classifier)
		}

		go func() {
			wg.Wait()

			close(errs)
		}()

		for err := range errs {
			if err != nil {
				return err
			}
		}

		return nil
	}

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

					d := lutil.LexiconTypeDecoder{
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
						createdAt, err := time.Parse(time.RFC3339Nano, post.CreatedAt)
						if err != nil {
							createdAt, err = time.Parse("2006-01-02T15:04:05.999999", post.CreatedAt) // For some reason, Bsky sometimes seems to not specify the timezone
							if err != nil {

								log.Println("Could not parse post date, skipping:", err)

								continue l
							}
						}

						post, err := p.CreatePost(
							ctx,
							rp.RepoDid(),
							path.Base(op.Path),
							createdAt,
							post.Text,
							post.Reply != nil,
							post.Langs,
						)
						if err != nil {
							log.Println("Could not insert post, skipping:", err)

							continue l
						}

						if *verbose {
							log.Println("Created post", post)
						}

						if err := classify(post); err != nil {
							panic(err)
						}
					} else if post.LexiconTypeID == "app.bsky.feed.like" {
						var like bsky.FeedLike
						if err := json.Unmarshal(b, &like); err != nil {
							log.Println("Could not unmarshal like, skipping:", err)

							continue l
						}

						u, err := iutil.ParseAtUri(like.Subject.Uri)
						if err != nil {
							log.Println("Could not parse like subject URI, skipping:", err)

							continue l
						}

						post, err := p.LikePost(
							ctx,
							u.Did,
							u.Rkey,
						)
						if err != nil {
							if !errors.Is(err, sql.ErrNoRows) {
								log.Println("Could not update post, skipping:", err)
							}

							continue l
						}

						if *verbose {
							log.Println("Liked post", post)
						}

						if err := classify(post); err != nil {
							panic(err)
						}
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
