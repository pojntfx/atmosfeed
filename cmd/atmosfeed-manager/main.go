package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/bluesky-social/indigo/api/atproto"
	"github.com/bluesky-social/indigo/api/bsky"
	"github.com/bluesky-social/indigo/events"
	"github.com/bluesky-social/indigo/events/schedulers/sequential"
	lutil "github.com/bluesky-social/indigo/lex/util"
	"github.com/bluesky-social/indigo/repo"
	"github.com/bluesky-social/indigo/repomgr"
	iutil "github.com/bluesky-social/indigo/util"
	"github.com/bluesky-social/indigo/xrpc"
	"github.com/gorilla/websocket"
	"github.com/pojntfx/atmosfeed/pkg/models"
	"github.com/pojntfx/atmosfeed/pkg/persisters"
	"github.com/redis/go-redis/v9"
)

const (
	lexiconFeedPost = "app.bsky.feed.post"
)

var (
	errMissingFeedURI         = errors.New("missing feed URI")
	errInvalidFeedURI         = errors.New("invalid feed URI")
	errInvalidLimit           = errors.New("invalid limit")
	errLimitTooHigh           = errors.New("limit too high")
	errInvalidFeedCursor      = errors.New("invalid feed cursor")
	errCouldNotEncode         = errors.New("could not encode")
	errCouldNotGetSession     = errors.New("could not get session")
	errCouldNotGetFeeds       = errors.New("could not get feeds")
	errMissingRkey            = errors.New("missing rkey")
	errCouldNotReadClassifier = errors.New("could not read classifier")
	errCouldNotUpsertFeed     = errors.New("could not upsert feed")
	errCouldNotDeleteFeed     = errors.New("could not delete feed")
)

type feedSkeleton struct {
	Feed   []feedSkeletonPost `json:"feed"`
	Cursor string             `json:"cursor"`
}

type feedSkeletonPost struct {
	Post string `json:"post"`
}

type wellKnownDidDocument struct {
	Context []string           `json:"@context"`
	ID      string             `json:"id"`
	Service []wellKnownService `json:"service"`
}

type wellKnownService struct {
	ID              string `json:"id"`
	Type            string `json:"type"`
	ServiceEndpoint string `json:"serviceEndpoint"`
}

func main() {

	pdsURL := flag.String("pds-url", "https://bsky.social", "PDS URL")
	postgresURL := flag.String("postgres-url", "postgresql://postgres@localhost:5432/atmosfeed?sslmode=disable", "PostgreSQL URL")
	redisURL := flag.String("redis-url", "redis://localhost:6379/0", "Redis URL")
	s3URL := flag.String("s3-url", "http://minioadmin:minioadmin@localhost:9000?bucket=atmosfeed", "S3 URL")

	laddr := flag.String("laddr", "localhost:1337", "Listen address")

	ttl := flag.Duration("ttl", time.Hour*6, "Maximum age of posts to return for a feed")
	limit := flag.Int("limit", 100, "Maximum amount of posts to return for a feed")

	feedGeneratorDID := flag.String("feed-generator-did", "did:web:atmosfeed-feeds.serveo.net", "DID of the feed generator (typically the hostname of the publicly reachable URL)")
	feedGeneratorURL := flag.String("feed-generator-url", "https://atmosfeed-feeds.serveo.net", "Publicly reachable URL of the feed generator")

	verbose := flag.Bool("verbose", false, "Whether to enable verbose logging")

	flag.Parse()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	pu, err := url.Parse(*pdsURL)
	if err != nil {
		panic(err)
	}
	pu.Scheme = "wss"
	pu = pu.JoinPath("xrpc", "com.atproto.sync.subscribeRepos")

	conn, _, err := websocket.DefaultDialer.DialContext(ctx, pu.String(), nil)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	log.Println("Connected to PDS", *pdsURL)

	options, err := redis.ParseURL(*redisURL)
	if err != nil {
		panic(err)
	}

	broker := redis.NewClient(options)
	defer broker.Close()

	log.Println("Connected to Redis")

	persister := persisters.NewManagerPersister(*postgresURL, broker, *s3URL)

	if err := persister.Init(ctx); err != nil {
		panic(err)
	}

	log.Println("Connected to PostgreSQL and S3")

	lis, err := net.Listen("tcp", *laddr)
	if err != nil {
		panic(err)
	}
	defer lis.Close()

	log.Println("Listening on", lis.Addr())

	mux := http.NewServeMux()

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
						if _, err := broker.XAdd(ctx, &redis.XAddArgs{
							Stream: persisters.StreamPostInsert,
							Values: map[string]interface{}{
								"did":       rp.RepoDid(),
								"rkey":      path.Base(op.Path),
								"createdAt": post.CreatedAt,
								"text":      post.Text,
								"reply":     post.Reply != nil,
								"langs":     strings.Join(post.Langs, ","),
							},
						}).Result(); err != nil {
							log.Println("Could not publish post, skipping:", err)

							continue l
						}

						if *verbose {
							log.Println("Published post", post)
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

						if _, err := broker.XAdd(ctx, &redis.XAddArgs{
							Stream: persisters.StreamPostLike,
							Values: map[string]interface{}{
								"did":  u.Did,
								"rkey": u.Rkey,
							},
						}).Result(); err != nil {
							log.Println("Could not publish like, skipping:", err)

							continue l
						}

						if *verbose {
							log.Println("Published like", post)
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
				http.Error(w, errMissingFeedURI.Error(), http.StatusUnprocessableEntity)

				log.Println(errMissingFeedURI)

				return
			}

			u, err := iutil.ParseAtUri(feedURL)
			if err != nil {
				http.Error(w, errInvalidFeedURI.Error(), http.StatusUnprocessableEntity)

				log.Println(errInvalidFeedURI)

				return
			}

			rawFeedLimit := r.URL.Query().Get("limit")
			if strings.TrimSpace(rawFeedLimit) == "" {
				rawFeedLimit = "1"
			}

			feedLimit, err := strconv.Atoi(rawFeedLimit)
			if err != nil {
				http.Error(w, errInvalidLimit.Error(), http.StatusUnprocessableEntity)

				log.Println(errInvalidLimit)

				return
			}

			if feedLimit > *limit {
				http.Error(w, errLimitTooHigh.Error(), http.StatusUnprocessableEntity)

				log.Println(errLimitTooHigh)

				return
			}

			feedCursor := r.URL.Query().Get("cursor")

			defer func() {
				if err := recover(); err != nil {
					w.WriteHeader(http.StatusInternalServerError)

					log.Printf("Client disconnected with error: %v", err)
				}
			}()

			rawFeedPosts := []models.GetFeedPostsRow{}
			if strings.TrimSpace(feedCursor) == "" {
				rawFeedPosts, err = persister.GetFeedPosts(
					ctx,
					u.Did,
					u.Rkey,
					time.Now().Add(-*ttl),
					int32(feedLimit),
				)
				if err != nil {
					panic(err)
				}
			} else {
				cursor, err := iutil.ParseAtUri(feedCursor)
				if err != nil {
					http.Error(w, errInvalidFeedCursor.Error(), http.StatusUnprocessableEntity)

					log.Println(errInvalidFeedCursor)

					return
				}

				fp, err := persister.GetFeedPostsCursor(
					ctx,
					u.Did,
					u.Rkey,
					time.Now().Add(-*ttl),
					int32(feedLimit),
					cursor.Did,
					cursor.Rkey,
				)
				if err != nil {
					panic(err)
				}

				for _, p := range fp {
					rawFeedPosts = append(rawFeedPosts, models.GetFeedPostsRow(p))
				}
			}

			res := feedSkeleton{
				Feed: []feedSkeletonPost{},
			}
			for _, rawFeedPost := range rawFeedPosts {
				res.Feed = append(res.Feed, feedSkeletonPost{
					Post: fmt.Sprintf("at://%s/%s/%s", rawFeedPost.Did, lexiconFeedPost, rawFeedPost.Rkey),
				})
			}

			if len(res.Feed) > 0 {
				res.Cursor = res.Feed[len(res.Feed)-1].Post
			}

			w.Header().Set("Content-Type", "application/json")

			if err := json.NewEncoder(w).Encode(res); err != nil {
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

			res := wellKnownDidDocument{
				Context: []string{"https://www.w3.org/ns/did/v1"},
				ID:      *feedGeneratorDID,
				Service: []wellKnownService{
					{
						ID:              "#bsky_fg",
						Type:            "BskyFeedGenerator",
						ServiceEndpoint: *feedGeneratorURL,
					},
				},
			}

			w.Header().Set("Content-Type", "application/json")

			if err := json.NewEncoder(w).Encode(res); err != nil {
				panic(fmt.Errorf("%w: %v", errCouldNotEncode, err))
			}
		}))

		mux.HandleFunc("/admin/feeds", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			accessJwt := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
			if strings.TrimSpace(accessJwt) == "" {
				w.WriteHeader(http.StatusUnauthorized)

				return
			}

			defer func() {
				if err := recover(); err != nil {
					w.WriteHeader(http.StatusInternalServerError)

					log.Printf("Client disconnected with error: %v", err)
				}
			}()

			client := &xrpc.Client{
				Client: http.DefaultClient,
				Host:   *pdsURL,
				Auth: &xrpc.AuthInfo{
					AccessJwt: accessJwt,
				},
			}

			session, err := atproto.ServerGetSession(r.Context(), client)
			if err != nil {
				panic(fmt.Errorf("%w: %v", errCouldNotGetSession, err))
			}

			switch r.Method {
			case http.MethodGet:
				rawAdminFeeds, err := persister.GetFeedsForDid(r.Context(), session.Did)
				if err != nil {
					panic(fmt.Errorf("%w: %v", errCouldNotGetFeeds, err))
				}

				res := []string{}
				for _, rawFeed := range rawAdminFeeds {
					res = append(res, rawFeed.Rkey)
				}

				w.Header().Set("Content-Type", "application/json")

				if err := json.NewEncoder(w).Encode(res); err != nil {
					panic(fmt.Errorf("%w: %v", errCouldNotEncode, err))
				}

			case http.MethodPut:
				rkey := r.URL.Query().Get("rkey")
				if strings.TrimSpace(rkey) == "" {
					http.Error(w, errMissingRkey.Error(), http.StatusUnprocessableEntity)

					log.Println(errMissingRkey)

					return
				}

				if err := persister.UpsertFeed(ctx, session.Did, rkey, r.Body); err != nil {
					panic(fmt.Errorf("%w: %v", errCouldNotUpsertFeed, err))
				}

			case http.MethodDelete:
				rkey := r.URL.Query().Get("rkey")
				if strings.TrimSpace(rkey) == "" {
					http.Error(w, errMissingRkey.Error(), http.StatusUnprocessableEntity)

					log.Println(errMissingRkey)

					return
				}

				if err := persister.DeleteFeed(ctx, session.Did, rkey); err != nil {
					panic(fmt.Errorf("%w: %v", errCouldNotDeleteFeed, err))
				}

			default:
				w.WriteHeader(http.StatusMethodNotAllowed)
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