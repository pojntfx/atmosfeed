package cmd

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
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
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	laddrFlag            = "laddr"
	ttlFlag              = "ttl"
	limitFlag            = "limit"
	feedGeneratorDIDFlag = "feed-generator-did"
	feedGeneratorURLFlag = "feed-generator-url"
	bgsURLFlag           = "bgs-url"

	lexiconFeedPost = "app.bsky.feed.post"

	originFlag         = "origin"
	deleteAllPostsFlag = "delete-all-posts"
)

var (
	errMissingFeedURI          = errors.New("missing feed URI")
	errInvalidFeedURI          = errors.New("invalid feed URI")
	errInvalidLimit            = errors.New("invalid limit")
	errLimitTooHigh            = errors.New("limit too high")
	errInvalidFeedCursor       = errors.New("invalid feed cursor")
	errCouldNotEncode          = errors.New("could not encode")
	errCouldNotGetSession      = errors.New("could not get session")
	errCouldNotGetFeeds        = errors.New("could not get feeds")
	errCouldNotGetPosts        = errors.New("could not get posts")
	errCouldNotGetFeedPosts    = errors.New("could not get feed posts")
	errMissingRkey             = errors.New("missing rkey")
	errCouldNotUpsertFeed      = errors.New("could not upsert feed")
	errCouldNotDeleteFeed      = errors.New("could not delete feed")
	errMissingService          = errors.New("missing service")
	errMissingResource         = errors.New("missing resource")
	errInvalidResource         = errors.New("invalid resource")
	errCouldNotDeletePosts     = errors.New("could not delete posts")
	errCouldNotDeleteFeedPosts = errors.New("could not delete feed posts")
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

type structuredUserdata struct {
	Feeds     []models.Feed     `json:"feeds"`
	Posts     []models.Post     `json:"posts"`
	FeedPosts []models.FeedPost `json:"feedPosts"`
}

var managerCmd = &cobra.Command{
	Use:     "manager",
	Aliases: []string{"m"},
	Short:   "Start an Atmosfeed manager",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := viper.BindPFlags(cmd.PersistentFlags()); err != nil {
			return err
		}

		pu, err := url.Parse(viper.GetString(bgsURLFlag))
		if err != nil {
			return err
		}
		pu.Scheme = "wss"
		pu = pu.JoinPath("xrpc", "com.atproto.sync.subscribeRepos")

		conn, _, err := websocket.DefaultDialer.DialContext(cmd.Context(), pu.String(), nil)
		if err != nil {
			return err
		}
		defer conn.Close()

		log.Println("Connected to PDS", viper.GetString(bgsURLFlag))

		options, err := redis.ParseURL(viper.GetString(redisURLFlag))
		if err != nil {
			return err
		}

		broker := redis.NewClient(options)
		defer broker.Close()

		log.Println("Connected to Redis")

		persister := persisters.NewManagerPersister(viper.GetString(postgresURLFlag), broker, viper.GetString(s3URLFlag))

		if err := persister.Init(cmd.Context()); err != nil {
			return err
		}

		log.Println("Connected to PostgreSQL and S3")

		if viper.GetBool(deleteAllPostsFlag) {
			if viper.GetBool(verboseFlag) {
				log.Println("Deleting all posts")
			}

			if err := persister.DeleteAllPosts(cmd.Context()); err != nil {
				return err
			}
		}

		lis, err := net.Listen("tcp", viper.GetString(laddrFlag))
		if err != nil {
			return err
		}
		defer lis.Close()

		log.Println("Listening on", lis.Addr())

		mux := http.NewServeMux()

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

			if feedLimit > viper.GetInt(limitFlag) {
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
					cmd.Context(),
					u.Did,
					u.Rkey,
					time.Now().Add(-viper.GetDuration(ttlFlag)),
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
					cmd.Context(),
					u.Did,
					u.Rkey,
					time.Now().Add(-viper.GetDuration(ttlFlag)),
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
				ID:      viper.GetString(feedGeneratorDIDFlag),
				Service: []wellKnownService{
					{
						ID:              "#bsky_fg",
						Type:            "BskyFeedGenerator",
						ServiceEndpoint: viper.GetString(feedGeneratorURLFlag),
					},
				},
			}

			w.Header().Set("Content-Type", "application/json")

			if err := json.NewEncoder(w).Encode(res); err != nil {
				panic(fmt.Errorf("%w: %v", errCouldNotEncode, err))
			}
		}))

		authorize := func(w http.ResponseWriter, r *http.Request) *atproto.ServerGetSession_Output {
			if o := r.Header.Get("Origin"); o == viper.GetString(originFlag) {
				w.Header().Set("Access-Control-Allow-Origin", o)
				w.Header().Set("Access-Control-Allow-Methods", "GET, PUT, PATCH, DELETE")
				w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
				w.Header().Set("Access-Control-Allow-Credentials", "true")
			}

			if r.Method == http.MethodOptions {
				return nil
			}

			accessJWT := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
			if strings.TrimSpace(accessJWT) == "" {
				w.WriteHeader(http.StatusUnauthorized)

				return nil
			}

			service := r.URL.Query().Get("service")
			if strings.TrimSpace(service) == "" {
				http.Error(w, errMissingService.Error(), http.StatusUnprocessableEntity)

				log.Println(errMissingService)

				return nil
			}

			client := &xrpc.Client{
				Client: http.DefaultClient,
				Host:   service,
				Auth: &xrpc.AuthInfo{
					AccessJwt: accessJWT,
				},
			}

			session, err := atproto.ServerGetSession(r.Context(), client)
			if err != nil {
				http.Error(w, errCouldNotGetSession.Error(), http.StatusUnprocessableEntity)

				log.Println(errCouldNotGetSession)

				return nil
			}

			return session
		}

		mux.HandleFunc("/admin/feeds", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			session := authorize(w, r)
			if session == nil {
				return
			}

			defer func() {
				if err := recover(); err != nil {
					w.WriteHeader(http.StatusInternalServerError)

					log.Printf("Client disconnected with error: %v", err)
				}
			}()

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

				pinnedDID := r.URL.Query().Get("pinnedDID")
				pinnedRkey := r.URL.Query().Get("pinnedRkey")

				if err := persister.UpsertFeed(cmd.Context(), session.Did, rkey, pinnedDID, pinnedRkey, r.Body); err != nil {
					panic(fmt.Errorf("%w: %v", errCouldNotUpsertFeed, err))
				}

			// Same as PUT, except we don't replace the classifier
			case http.MethodPatch:
				rkey := r.URL.Query().Get("rkey")
				if strings.TrimSpace(rkey) == "" {
					http.Error(w, errMissingRkey.Error(), http.StatusUnprocessableEntity)

					log.Println(errMissingRkey)

					return
				}

				pinnedDID := r.URL.Query().Get("pinnedDID")
				pinnedRkey := r.URL.Query().Get("pinnedRkey")

				if err := persister.UpsertFeed(cmd.Context(), session.Did, rkey, pinnedDID, pinnedRkey, nil); err != nil {
					panic(fmt.Errorf("%w: %v", errCouldNotUpsertFeed, err))
				}

			case http.MethodDelete:
				rkey := r.URL.Query().Get("rkey")
				if strings.TrimSpace(rkey) == "" {
					http.Error(w, errMissingRkey.Error(), http.StatusUnprocessableEntity)

					log.Println(errMissingRkey)

					return
				}

				if err := persister.DeleteFeed(cmd.Context(), session.Did, rkey); err != nil {
					panic(fmt.Errorf("%w: %v", errCouldNotDeleteFeed, err))
				}

			default:
				w.WriteHeader(http.StatusMethodNotAllowed)
			}
		}))

		mux.HandleFunc("/userdata", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			session := authorize(w, r)
			if session == nil {
				return
			}

			defer func() {
				if err := recover(); err != nil {
					w.WriteHeader(http.StatusInternalServerError)

					log.Printf("Client disconnected with error: %v", err)
				}
			}()

			switch r.Method {
			case http.MethodDelete:
				feeds, err := persister.GetFeedsForDid(r.Context(), session.Did)
				if err != nil {
					panic(fmt.Errorf("%w: %v", errCouldNotGetFeeds, err))
				}

				for _, feed := range feeds {
					if err := persister.DeleteFeed(r.Context(), session.Did, feed.Rkey); err != nil {
						panic(fmt.Errorf("%w: %v", errCouldNotDeleteFeed, err))
					}
				}

				if err := persister.DeletePostsForDid(r.Context(), session.Did); err != nil {
					panic(fmt.Errorf("%w: %v", errCouldNotDeletePosts, err))
				}

				if err := persister.DeleteFeedPostsForDid(r.Context(), session.Did); err != nil {
					panic(fmt.Errorf("%w: %v", errCouldNotDeleteFeedPosts, err))
				}

			default:
				w.WriteHeader(http.StatusMethodNotAllowed)
			}
		}))

		mux.HandleFunc("/userdata/structured", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			session := authorize(w, r)
			if session == nil {
				return
			}

			defer func() {
				if err := recover(); err != nil {
					w.WriteHeader(http.StatusInternalServerError)

					log.Printf("Client disconnected with error: %v", err)
				}
			}()

			switch r.Method {
			case http.MethodGet:
				feeds, err := persister.GetFeedsForDid(r.Context(), session.Did)
				if err != nil {
					panic(fmt.Errorf("%w: %v", errCouldNotGetFeeds, err))
				}

				posts, err := persister.GetPostsForDid(r.Context(), session.Did)
				if err != nil {
					panic(fmt.Errorf("%w: %v", errCouldNotGetPosts, err))
				}

				feedPosts, err := persister.GetFeedPostsForDid(r.Context(), session.Did)
				if err != nil {
					panic(fmt.Errorf("%w: %v", errCouldNotGetFeedPosts, err))
				}

				w.Header().Set("Content-Type", "application/json")

				if err := json.NewEncoder(w).Encode(structuredUserdata{
					Feeds:     feeds,
					Posts:     posts,
					FeedPosts: feedPosts,
				}); err != nil {
					panic(fmt.Errorf("%w: %v", errCouldNotEncode, err))
				}

			default:
				w.WriteHeader(http.StatusMethodNotAllowed)
			}
		}))

		mux.HandleFunc("/userdata/blob", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			session := authorize(w, r)
			if session == nil {
				return
			}

			defer func() {
				if err := recover(); err != nil {
					w.WriteHeader(http.StatusInternalServerError)

					log.Printf("Client disconnected with error: %v", err)
				}
			}()

			resource := r.URL.Query().Get("resource")
			if strings.TrimSpace(resource) == "" {
				panic(fmt.Errorf("%w: %v", errMissingResource, err))
			}

			if resource != "classifier" {
				panic(fmt.Errorf("%w: %v", errInvalidResource, err))
			}

			rkey := r.URL.Query().Get("rkey")
			if strings.TrimSpace(rkey) == "" {
				panic(fmt.Errorf("%w: %v", errMissingRkey, err))
			}

			switch r.Method {
			case http.MethodGet:
				classifier, err := persister.GetFeedClassifier(r.Context(), session.Did, rkey)
				if err != nil {
					panic(fmt.Errorf("%w: %v", errCouldNotGetFeedPosts, err))
				}

				w.Header().Set("Content-Type", "application/octet-stream")

				if _, err := io.Copy(w, classifier); err != nil {
					panic(fmt.Errorf("%w: %v", errCouldNotEncode, err))
				}

			default:
				w.WriteHeader(http.StatusMethodNotAllowed)
			}
		}))

		handlers := events.RepoStreamCallbacks{
			RepoCommit: func(c *atproto.SyncSubscribeRepos_Commit) error {
				rp, err := repo.ReadRepoFromCar(cmd.Context(), bytes.NewReader(c.Blocks))
				if err != nil {
					log.Println("Could not parse repo, skipping:", err)

					return nil
				}

			l:
				for _, op := range c.Ops {
					switch repomgr.EventKind(op.Action) {
					case repomgr.EvtKindCreateRecord:
						_, res, err := rp.GetRecord(cmd.Context(), op.Path)
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
							if _, err := broker.XAdd(cmd.Context(), &redis.XAddArgs{
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

							if viper.GetBool(verboseFlag) {
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

							if _, err := broker.XAdd(cmd.Context(), &redis.XAddArgs{
								Stream: persisters.StreamPostLike,
								Values: map[string]interface{}{
									"did":  u.Did,
									"rkey": u.Rkey,
								},
							}).Result(); err != nil {
								log.Println("Could not publish like, skipping:", err)

								continue l
							}

							if viper.GetBool(verboseFlag) {
								log.Println("Published like", post)
							}
						}

					case repomgr.EvtKindDeleteRecord:
						if lexiconTypeID := path.Dir(op.Path); lexiconTypeID == lexiconFeedPost {
							did, rkey := rp.SignedCommit().Did, path.Base(op.Path)
							if err := persister.DeletePost(cmd.Context(), did, rkey); err != nil {
								log.Println("Could not delete post, skipping:", err)

								continue l
							}

							if viper.GetBool(verboseFlag) {
								log.Println("Deleted post", did, rkey)
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
				cmd.Context(),
				conn,
				sequential.NewScheduler(
					conn.RemoteAddr().String(),
					handlers.EventHandler,
				),
			); err != nil {
				errs <- err

				return
			}
		}()

		go func() {
			if err := http.Serve(lis, mux); err != nil {
				errs <- err

				return
			}
		}()

		for err := range errs {
			if err != nil {
				return err
			}
		}

		return nil
	},
}

func init() {
	managerCmd.PersistentFlags().String(bgsURLFlag, "https://bsky.network", "BGS URL")
	managerCmd.PersistentFlags().String(laddrFlag, "localhost:1337", "Listen address")
	managerCmd.PersistentFlags().Duration(ttlFlag, time.Hour*6, "Maximum age of posts to return for a feed")
	managerCmd.PersistentFlags().Int(limitFlag, 100, "Maximum amount of posts to return for a feed")
	managerCmd.PersistentFlags().String(feedGeneratorDIDFlag, "did:web:atmosfeed-feeds.serveo.net", "DID of the feed generator (typically the hostname of the publicly reachable URL)")
	managerCmd.PersistentFlags().String(feedGeneratorURLFlag, "https://atmosfeed-feeds.serveo.net", "Publicly reachable URL of the feed generator")
	managerCmd.PersistentFlags().String(originFlag, "https://atmosfeed.p8.lu", "Allowed CORS origin")
	managerCmd.PersistentFlags().Bool(deleteAllPostsFlag, true, "Whether to delete all posts from the index on startup (required for compliance with the EU right to be forgotten/GDPR article 17; deletions during uptime are handled using delete commits)")

	viper.AutomaticEnv()

	rootCmd.AddCommand(managerCmd)
}
