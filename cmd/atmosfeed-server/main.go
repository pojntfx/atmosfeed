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
	"net"
	"net/http"
	"net/url"
	"os"
	"path"
	"signature"
	"strings"
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

	errPostgresForeignKeyViolation = "23503"

	lexiconFeedPost = "app.bsky.feed.post"
)

var (
	errMissingFeed    = errors.New("missing feed")
	errInvalidFeedURI = errors.New("invalid feed URI")
	errCouldNotEncode = errors.New("could not encode")
)

type feed struct {
	Feed   []feedPost `json:"feed"`
	Cursor string     `json:"cursor"`
}

type feedPost struct {
	Post string `json:"post"`
}

type didDocument struct {
	Context []string  `json:"@context"`
	ID      string    `json:"id"`
	Service []service `json:"service"`
}

type service struct {
	ID              string `json:"id"`
	Type            string `json:"type"`
	ServiceEndpoint string `json:"serviceEndpoint"`
}

func main() {
	pdsURL := flag.String("pds-url", "wss://bsky.social/", "PDS URL (can also be set using `PDS_URL` env variable)")
	postgresURL := flag.String("postgres-url", "postgresql://postgres@localhost:5432/atmosfeed?sslmode=disable", "PostgreSQL URL (can also be set using `POSTGRES_URL` env variable)")
	verbose := flag.Bool("verbose", false, "Whether to enable verbose logging")
	laddr := flag.String("laddr", "localhost:1337", "Listen address")
	classifierTimeout := flag.Duration("classifier-timeout", time.Second, "Amount of time after which to stop a classifer Scale function from running")
	ttl := flag.Duration("ttl", time.Hour*6, "Maximum age of posts to return for a feed")
	limit := flag.Int("limit", 100, "Maximum amount of posts to return for a feed")
	did := flag.String("did", "did:web:atmosfeed-feeds.serveo.net", "DID to publish the feed generator under")
	serviceEndpoint := flag.String("service-endpoint", "https://atmosfeed-feeds.serveo.net", "Publicly reachable endpoint for the feed generator")

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

	persister := persisters.NewPersister(*postgresURL)

	if err := persister.Init(); err != nil {
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
					classifierSource, err := persister.GetFeedClassifier(ctx, notification.Extra)
					if err != nil {
						log.Println("Could not fetch new classifier, skipping:", err)

						return
					}

					classifierLock.Lock()
					defer classifierLock.Unlock()

					fn := &scalefunc.Schema{}
					if err := fn.Decode(classifierSource); err != nil {
						log.Println("Could not parse classifier, skipping:", err)

						return
					}

					runtime, err := scale.New(scale.NewConfig(signature.New).WithFunction(fn))
					if err != nil {
						log.Println("Could not start classifier runtime, skipping:", err)

						return
					}

					instance, err := runtime.Instance()
					if err != nil {
						log.Println("Could not start classifier instance, skipping:", err)

						return
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

	lis, err := net.Listen("tcp", *laddr)
	if err != nil {
		panic(err)
	}
	defer lis.Close()

	log.Println("Listening on", lis.Addr())

	mux := http.NewServeMux()

	classifierSources, err := persister.GetFeeds(ctx)
	if err != nil {
		panic(err)
	}

	for _, classifierSource := range classifierSources {
		func() {
			classifierLock.Lock()
			defer classifierLock.Unlock()

			fn := &scalefunc.Schema{}
			if err := fn.Decode(classifierSource.Classifier); err != nil {
				log.Println("Could not parse classifier, skipping:", err)

				return
			}

			runtime, err := scale.New(scale.NewConfig(signature.New).WithFunction(fn))
			if err != nil {
				log.Println("Could not start classifier runtime, skipping:", err)

				return
			}

			instance, err := runtime.Instance()
			if err != nil {
				log.Println("Could not start classifier instance, skipping:", err)

				return
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

			go func(feedName string, classifier *scale.Instance[*signature.Signature]) {
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
					if err := persister.CreateFeedPost(ctx, feedName, p.Did, p.Rkey); err != nil {
						// We can safely ignore inserts if the feed that it should be inserted in was deleted
						if err, ok := err.(*pq.Error); !ok || err.Code != pq.ErrorCode(errPostgresForeignKeyViolation) {
							errs <- err
						}

						return
					}
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

					if post.LexiconTypeID == lexiconFeedPost {
						createdAt, err := time.Parse(time.RFC3339Nano, post.CreatedAt)
						if err != nil {
							createdAt, err = time.Parse("2006-01-02T15:04:05.999999", post.CreatedAt) // For some reason, Bsky sometimes seems to not specify the timezone
							if err != nil {

								log.Println("Could not parse post date, skipping:", err)

								continue l
							}
						}

						post, err := persister.CreatePost(
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
							log.Println("Could not classify post, skipping:", err)

							continue l
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

						post, err := persister.LikePost(
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
							return err
						}
					}
				}
			}

			return nil
		},
	}

	errs := make(chan error)

	go func() {
		if err := events.HandleRepoStream(
			ctx,
			conn,
			sequential.NewScheduler(
				conn.RemoteAddr().String(),
				handlers.EventHandler,
			),
		); err != nil {
			errs <- err
		}
	}()

	go func() {
		mux.HandleFunc("/xrpc/app.bsky.feed.getFeedSkeleton", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			feedURL := r.URL.Query().Get("feed")
			if strings.TrimSpace(feedURL) == "" {
				http.Error(w, errMissingFeed.Error(), http.StatusUnprocessableEntity)

				log.Println(errMissingFeed)

				return
			}

			u, err := iutil.ParseAtUri(feedURL)
			if err != nil {
				http.Error(w, errInvalidFeedURI.Error(), http.StatusUnprocessableEntity)

				log.Println(errInvalidFeedURI)

				return
			}

			defer func() {
				if err := recover(); err != nil {
					w.WriteHeader(http.StatusInternalServerError)

					log.Printf("Client disconnected with error: %v", err)
				}
			}()

			rawFeedPosts, err := persister.GetFeedPosts(ctx, u.Rkey, time.Now().Add(-*ttl), int32(*limit))
			if err != nil {
				panic(err)
			}

			feedResponse := feed{
				Feed: []feedPost{},
			}
			for _, rawFeedPost := range rawFeedPosts {
				feedResponse.Feed = append(feedResponse.Feed, feedPost{
					Post: fmt.Sprintf("at://%s/%s/%s", rawFeedPost.Did, lexiconFeedPost, rawFeedPost.Rkey),
				})
			}

			w.Header().Set("Content-Type", "application/json")

			if err := json.NewEncoder(w).Encode(feedResponse); err != nil {
				panic(fmt.Errorf("%w: %v", errCouldNotEncode, err))
			}
		}))

		mux.HandleFunc("/.well-known/did.json", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					w.WriteHeader(http.StatusInternalServerError)

					log.Printf("Client disconnected with error: %v", err)
				}
			}()

			didResponse := didDocument{
				Context: []string{"https://www.w3.org/ns/did/v1"},
				ID:      *did,
				Service: []service{
					{
						ID:              "#bsky_fg",
						Type:            "BskyFeedGenerator",
						ServiceEndpoint: *serviceEndpoint,
					},
				},
			}

			w.Header().Set("Content-Type", "application/json")

			if err := json.NewEncoder(w).Encode(didResponse); err != nil {
				panic(fmt.Errorf("%w: %v", errCouldNotEncode, err))
			}
		}))

		if err := http.Serve(lis, mux); err != nil {
			errs <- err

			return
		}
	}()

	for err := range errs {
		if err == nil {
			return
		}

		panic(err)
	}
}
