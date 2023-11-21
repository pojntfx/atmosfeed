package cmd

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/url"
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
	"github.com/loopholelabs/scale"
	"github.com/loopholelabs/scale/scalefunc"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	bgsURLFlag      = "bgs-url"
	frontendURLFlag = "frontend-url"

	verboseFlag = "verbose"
	quietFlag   = "quiet"

	minWeightFlag = "min-weight"
	maxPostsFlag  = "max-posts"

	lexiconFeedPost = "app.bsky.feed.post"
)

var (
	errMessageInvalidCreatedAt = errors.New("message contained invalid createdAt")
)

var devCmd = &cobra.Command{
	Use:     "dev",
	Aliases: []string{"v"},
	Short:   "Develop a feed classifier locally",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := viper.BindPFlags(cmd.PersistentFlags()); err != nil {
			return err
		}

		fn, err := scalefunc.Read(viper.GetString(feedClassifierFlag))
		if err != nil {
			return err
		}

		runtime, err := scale.New(scale.NewConfig(signature.New).WithFunction(fn))
		if err != nil {
			return err
		}

		classifier, err := runtime.Instance()
		if err != nil {
			return err
		}

		fu, err := url.Parse(viper.GetString(frontendURLFlag))
		if err != nil {
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

		log.Println("Connected to BGS", viper.GetString(bgsURLFlag))

		var postsLock sync.Mutex
		posts := map[string]*signature.Post{}
		postsCh := make(chan signature.Post)

		handlers := events.RepoStreamCallbacks{
			RepoCommit: func(c *atproto.SyncSubscribeRepos_Commit) error {
				rp, err := repo.ReadRepoFromCar(cmd.Context(), bytes.NewReader(c.Blocks))
				if err != nil {
					if !viper.GetBool(quietFlag) {
						log.Println("Could not parse repo, skipping:", err)
					}

					return nil
				}

			l:
				for _, op := range c.Ops {
					switch repomgr.EventKind(op.Action) {
					case repomgr.EvtKindCreateRecord:
						_, res, err := rp.GetRecord(cmd.Context(), op.Path)
						if err != nil {
							if !viper.GetBool(quietFlag) {
								log.Println("Could not parse record, skipping:", err)
							}

							continue l
						}

						d := lutil.LexiconTypeDecoder{
							Val: res,
						}

						b, err := d.MarshalJSON()
						if err != nil {
							if !viper.GetBool(quietFlag) {
								log.Println("Could not marshal lexicon, skipping:", err)
							}

							continue l
						}

						var post bsky.FeedPost
						if err := json.Unmarshal(b, &post); err != nil {
							if !viper.GetBool(quietFlag) {
								log.Println("Could not unmarshal post, skipping:", err)
							}

							continue l
						}

						if post.LexiconTypeID == lexiconFeedPost {
							p := signature.NewPost()

							p.Did = rp.RepoDid()
							p.Rkey = path.Base(op.Path)
							p.Text = post.Text

							p.Langs = post.Langs

							createdAt, err := time.Parse(time.RFC3339Nano, post.CreatedAt)
							if err != nil {
								createdAt, err = time.Parse("2006-01-02T15:04:05.999999", post.CreatedAt) // For some reason, Bsky sometimes seems to not specify the timezone
								if err != nil {
									if !viper.GetBool(quietFlag) {
										log.Println(errMessageInvalidCreatedAt)
									}

									continue l
								}
							}

							p.CreatedAt = createdAt.Unix()
							p.Likes = 0

							p.Reply = post.Reply != nil

							postsLock.Lock()
							if len(posts) > viper.GetInt(maxPostsFlag) {
								posts = map[string]*signature.Post{}
							}
							posts[p.Did+"/"+p.Rkey] = p
							postsLock.Unlock()

							postsCh <- *p

							if viper.GetBool(verboseFlag) {
								log.Println("Published post", post)
							}
						} else if post.LexiconTypeID == "app.bsky.feed.like" {
							var like bsky.FeedLike
							if err := json.Unmarshal(b, &like); err != nil {
								if !viper.GetBool(quietFlag) {
									log.Println("Could not unmarshal like, skipping:", err)
								}

								continue l
							}

							u, err := iutil.ParseAtUri(like.Subject.Uri)
							if err != nil {
								if !viper.GetBool(quietFlag) {
									log.Println("Could not parse like subject URI, skipping:", err)
								}

								continue l
							}

							var p *signature.Post
							postsLock.Lock()
							po, ok := posts[u.Did+"/"+u.Rkey]
							if !ok {
								postsLock.Unlock()

								continue l
							}
							po.Likes++
							p = po
							postsLock.Unlock()

							postsCh <- *p

							if viper.GetBool(verboseFlag) {
								log.Println("Published like", post)
							}
						}
					}
				}

				return nil
			},
		}

		go func() {
			if err := events.HandleRepoStream(
				cmd.Context(),
				conn,
				sequential.NewScheduler(
					conn.RemoteAddr().String(),
					handlers.EventHandler,
				),
			); err != nil {
				panic(err)
			}
		}()

		for post := range postsCh {
			s := signature.New()
			s.Context.Post = &post

			if err := classifier.Run(cmd.Context(), s); err != nil {
				return err
			}

			if s.Context.Weight >= viper.GetInt64(minWeightFlag) {
				fmt.Println(s.Context.Weight, fu.JoinPath("profile", post.Did, "post", post.Rkey), post)
			}
		}

		return nil
	},
}

func init() {
	devCmd.PersistentFlags().String(feedClassifierFlag, "local-trending-latest.scale", "Path to the feed classifier to test")

	devCmd.PersistentFlags().String(bgsURLFlag, "https://bsky.network", "BGS URL")
	devCmd.PersistentFlags().String(frontendURLFlag, "https://bsky.app", "Bluesky frontend URL to use when logging posts")

	devCmd.PersistentFlags().Bool(verboseFlag, false, "Whether to enable verbose logging")
	devCmd.PersistentFlags().Bool(quietFlag, true, "Whether to silently ignore any non-fatal errors")

	devCmd.PersistentFlags().Int64(minWeightFlag, 0, "Minimum weight value the classifier has to return for a post to log it")
	devCmd.PersistentFlags().Int(maxPostsFlag, 1024*1024, "Maximum amount of posts to store in memory before clearing the cache")

	viper.AutomaticEnv()

	rootCmd.AddCommand(devCmd)
}
