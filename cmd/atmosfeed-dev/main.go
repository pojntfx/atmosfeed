package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"flag"
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
)

const (
	lexiconFeedPost = "app.bsky.feed.post"
)

var (
	errMessageInvalidCreatedAt = errors.New("message contained invalid createdAt")
)

func main() {
	feedClassifier := flag.String("feed-classifier", "local-trending-latest.scale", "Path to the feed classifier to test")
	frontendURL := flag.String("frontend-url", "https://bsky.app", "Frontend URL to use when logging posts")
	bgsURL := flag.String("bgs-url", "https://bsky.network", "BGS URL")
	verbose := flag.Bool("verbose", false, "Whether to enable verbose logging")
	minimumWeight := flag.Int64("minimum-weight", 0, "Minimum weight value the classifier has to return for a post to log it")
	quiet := flag.Bool("quiet", true, "Whether to silently ignore any non-fatal decoding errors")
	maxPosts := flag.Int("max-posts", 1024*1024, "Maximum amount of posts to store in memory before clearing the cache")

	flag.Parse()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	fn, err := scalefunc.Read(*feedClassifier)
	if err != nil {
		panic(err)
	}

	runtime, err := scale.New(scale.NewConfig(signature.New).WithFunction(fn))
	if err != nil {
		panic(err)
	}

	classifier, err := runtime.Instance()
	if err != nil {
		panic(err)
	}

	fu, err := url.Parse(*frontendURL)
	if err != nil {
		panic(err)
	}

	pu, err := url.Parse(*bgsURL)
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

	log.Println("Connected to PDS", *bgsURL)

	var postsLock sync.Mutex
	posts := map[string]*signature.Post{}
	postsCh := make(chan signature.Post)

	handlers := events.RepoStreamCallbacks{
		RepoCommit: func(c *atproto.SyncSubscribeRepos_Commit) error {
			rp, err := repo.ReadRepoFromCar(ctx, bytes.NewReader(c.Blocks))
			if err != nil {
				if !*quiet {
					log.Println("Could not parse repo, skipping:", err)
				}

				return nil
			}

		l:
			for _, op := range c.Ops {
				switch repomgr.EventKind(op.Action) {
				case repomgr.EvtKindCreateRecord:
					_, res, err := rp.GetRecord(ctx, op.Path)
					if err != nil {
						if !*quiet {
							log.Println("Could not parse record, skipping:", err)
						}

						continue l
					}

					d := lutil.LexiconTypeDecoder{
						Val: res,
					}

					b, err := d.MarshalJSON()
					if err != nil {
						if !*quiet {
							log.Println("Could not marshal lexicon, skipping:", err)
						}

						continue l
					}

					var post bsky.FeedPost
					if err := json.Unmarshal(b, &post); err != nil {
						if !*quiet {
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
								if !*quiet {
									log.Println(errMessageInvalidCreatedAt)
								}

								continue l
							}
						}

						p.CreatedAt = createdAt.Unix()
						p.Likes = 0

						p.Reply = post.Reply != nil

						postsLock.Lock()
						if len(posts) > *maxPosts {
							posts = map[string]*signature.Post{}
						}
						posts[p.Did+"/"+p.Rkey] = p
						postsLock.Unlock()

						postsCh <- *p

						if *verbose {
							log.Println("Published post", post)
						}
					} else if post.LexiconTypeID == "app.bsky.feed.like" {
						var like bsky.FeedLike
						if err := json.Unmarshal(b, &like); err != nil {
							if !*quiet {
								log.Println("Could not unmarshal like, skipping:", err)
							}

							continue l
						}

						u, err := iutil.ParseAtUri(like.Subject.Uri)
						if err != nil {
							if !*quiet {
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

						if *verbose {
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
			ctx,
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

		if err := classifier.Run(ctx, s); err != nil {
			panic(err)
		}

		if s.Context.Weight >= *minimumWeight {
			fmt.Println(s.Context.Weight, fu.JoinPath("profile", post.Did, "post", post.Rkey), post)
		}
	}
}
