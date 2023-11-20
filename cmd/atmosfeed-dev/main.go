package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"flag"
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
)

const (
	lexiconFeedPost = "app.bsky.feed.post"
)

var (
	errMessageInvalidCreatedAt = errors.New("message contained invalid createdAt")
)

func main() {
	// feedClassifier := flag.String("feed-classifier", "out/local-trending-latest.scale", "Path to the feed classifier to test")
	bgsURL := flag.String("bgs-url", "https://bsky.network", "BGS URL")
	verbose := flag.Bool("verbose", false, "Whether to enable verbose logging")

	flag.Parse()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

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
						p := signature.NewPost()

						p.Did = rp.RepoDid()
						p.Rkey = path.Base(op.Path)
						p.Text = post.Text

						p.Langs = post.Langs

						createdAt, err := time.Parse(time.RFC3339Nano, post.CreatedAt)
						if err != nil {
							createdAt, err = time.Parse("2006-01-02T15:04:05.999999", post.CreatedAt) // For some reason, Bsky sometimes seems to not specify the timezone
							if err != nil {
								log.Println(errMessageInvalidCreatedAt)

								continue l
							}
						}

						p.CreatedAt = createdAt.Unix()
						p.Likes = 0

						p.Reply = post.Reply != nil

						postsLock.Lock()
						posts[p.Did+"/"+p.Rkey] = p
						postsLock.Unlock()

						postsCh <- *p

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
		log.Println(post)
	}
}
