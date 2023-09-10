package main

import (
	"context"
	"errors"
	"flag"
	"log"
	"os"
	"path"
	"path/filepath"
	"signature"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/loopholelabs/scale"
	"github.com/loopholelabs/scale/scalefunc"
	"github.com/pojntfx/atmosfeed/pkg/models"
	"github.com/pojntfx/atmosfeed/pkg/persisters"
	"github.com/redis/go-redis/v9"
)

const (
	classifiersPath = "classifiers"
)

var (
	errMessageMissingDID       = errors.New("message did not contain DID")
	errMessageInvalidDID       = errors.New("message contained invalid DID")
	errMessageMissingRkey      = errors.New("message did not contain rkey")
	errMessageInvalidRkey      = errors.New("message contained invalid rkey")
	errMessageMissingCreatedAt = errors.New("message did not contain createdAt")
	errMessageInvalidCreatedAt = errors.New("message contained invalid createdAt")
	errMessageMissingText      = errors.New("message did not contain text")
	errMessageInvalidText      = errors.New("message contained invalid text")
	errMessageMissingReply     = errors.New("message did not contain reply")
	errMessageInvalidReply     = errors.New("message contained invalid reply")
	errMessageMissingLangs     = errors.New("message did not contain langs")
	errMessageInvalidLangs     = errors.New("message contained invalid langs")

	errPostgresForeignKeyViolation = "23503"
)

func main() {
	home, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}

	postgresURL := flag.String("postgres-url", "postgresql://postgres@localhost:5432/atmosfeed?sslmode=disable", "PostgreSQL URL")
	redisURL := flag.String("redis-url", "redis://localhost:6379/0", "Redis URL")

	classifierTimeout := flag.Duration("classifier-timeout", time.Second, "Amount of time after which to stop a classifer Scale function from running")

	workingDirectory := flag.String("working-directory", filepath.Join(home, ".local", "share", "atmosfeed", "var", "lib", "atmosfeed"), "Working directory to use")

	verbose := flag.Bool("verbose", false, "Whether to enable verbose logging")

	flag.Parse()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := os.RemoveAll(filepath.Join(*workingDirectory, classifiersPath)); err != nil {
		panic(err)
	}

	options, err := redis.ParseURL(*redisURL)
	if err != nil {
		panic(err)
	}

	broker := redis.NewClient(options)
	defer broker.Close()

	log.Println("Connected to Redis")

	persister := persisters.NewWorkerPersister(*postgresURL, broker)

	if err := persister.Init(); err != nil {
		panic(err)
	}

	log.Println("Connected to PostgreSQL")

	var classifierLock sync.Mutex
	classifiers := map[string]*scale.Instance[*signature.Signature]{}

	errs := make(chan error)

	go func() {
		for {
			streams, err := broker.XReadGroup(ctx, &redis.XReadGroupArgs{
				Group:    persisters.StreamFeedUpsert,
				Consumer: uuid.NewString(),
				Streams:  []string{persisters.StreamFeedUpsert, ">"},
				Block:    0,
				Count:    10,
			}).Result()
			if err != nil {
				log.Println("Could not subscribe to feed upsert stream, skipping:", err)

				continue
			}

			for _, stream := range streams {
				for _, message := range stream.Messages {
					rawDid, ok := message.Values["did"]
					if !ok {
						errs <- errMessageMissingDID

						return
					}

					did, ok := rawDid.(string)
					if !ok {
						errs <- errMessageInvalidDID

						return
					}

					rawRkey, ok := message.Values["rkey"]
					if !ok {
						errs <- errMessageMissingRkey

						return
					}

					rkey, ok := rawRkey.(string)
					if !ok {
						errs <- errMessageInvalidRkey

						return
					}

					if *verbose {
						log.Println("Upserted feed", did, rkey)
					}

					func() {
						classifierSource, err := persister.GetFeedClassifier(ctx, did, rkey)
						if err != nil {
							log.Println("Could not fetch new classifier, skipping:", err)

							return
						}

						classifierPath := filepath.Join(*workingDirectory, classifiersPath, did, rkey)
						if err := os.MkdirAll(filepath.Dir(classifierPath), os.ModePerm); err != nil {
							log.Println("Could not prepare directory for classifier, skipping:", err)

							return
						}

						classifierLock.Lock()
						defer classifierLock.Unlock()

						if err := os.WriteFile(classifierPath, classifierSource, os.ModePerm); err != nil {
							log.Println("Could not write classifier to disk, skipping:", err)

							return
						}

						fn, err := scalefunc.Read(classifierPath)
						if err != nil {
							log.Println("Could not read classifier, skipping:", err)

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

						classifiers[path.Join(did, rkey)] = instance
					}()
				}
			}
		}
	}()

	go func() {
		for {
			streams, err := broker.XReadGroup(ctx, &redis.XReadGroupArgs{
				Group:    persisters.StreamFeedDelete,
				Consumer: uuid.NewString(),
				Streams:  []string{persisters.StreamFeedDelete, ">"},
				Block:    0,
				Count:    10,
			}).Result()
			if err != nil {
				log.Println("Could not subscribe to feed delete stream, skipping:", err)

				continue
			}

			for _, stream := range streams {
				for _, message := range stream.Messages {
					rawDid, ok := message.Values["did"]
					if !ok {
						errs <- errMessageMissingDID

						return
					}

					did, ok := rawDid.(string)
					if !ok {
						errs <- errMessageInvalidDID

						return
					}

					rawRkey, ok := message.Values["rkey"]
					if !ok {
						errs <- errMessageMissingRkey

						return
					}

					rkey, ok := rawRkey.(string)
					if !ok {
						errs <- errMessageInvalidRkey

						return
					}

					if *verbose {
						log.Println("Deleted feed", did, rkey)
					}

					classifierLock.Lock()
					defer classifierLock.Unlock()

					if err := os.Remove(filepath.Join(*workingDirectory, classifiersPath, did, rkey)); err != nil {
						log.Println("Could not remove classifier from disk, skipping:", err)

						continue
					}

					delete(classifiers, path.Join(did, rkey))
				}
			}
		}
	}()

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

			classifiers[path.Join(classifierSource.Did, classifierSource.Rkey)] = instance
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

			did, rkey := path.Dir(feed), path.Base(feed)

			go func(feedDid, feedRkey string, classifier *scale.Instance[*signature.Signature]) {
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

				if s.Context.Weight >= 0 {
					if err := persister.UpsertFeedPost(ctx, feedDid, feedRkey, p.Did, p.Rkey, int32(s.Context.Weight)); err != nil {
						// We can safely ignore inserts if the feed that it should be inserted in was deleted
						if err, ok := err.(*pq.Error); !ok || err.Code != pq.ErrorCode(errPostgresForeignKeyViolation) {
							errs <- err
						}

						return
					}
				}
			}(did, rkey, classifier)
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

	go func() {
		for {
			streams, err := broker.XReadGroup(ctx, &redis.XReadGroupArgs{
				Group:    persisters.StreamPostInsert,
				Consumer: uuid.NewString(),
				Streams:  []string{persisters.StreamPostInsert, ">"},
				Block:    0,
				Count:    10,
			}).Result()
			if err != nil {
				log.Println("Could not subscribe to post insert stream, skipping:", err)

				continue
			}

			for _, stream := range streams {
				for _, message := range stream.Messages {
					rawDid, ok := message.Values["did"]
					if !ok {
						errs <- errMessageMissingDID

						return
					}

					did, ok := rawDid.(string)
					if !ok {
						errs <- errMessageInvalidDID

						return
					}

					rawRkey, ok := message.Values["rkey"]
					if !ok {
						errs <- errMessageMissingRkey

						return
					}

					rkey, ok := rawRkey.(string)
					if !ok {
						errs <- errMessageInvalidRkey

						return
					}

					rawCreatedAt, ok := message.Values["createdAt"]
					if !ok {
						errs <- errMessageMissingCreatedAt

						return
					}

					createdAtRFC, ok := rawCreatedAt.(string)
					if !ok {
						errs <- errMessageInvalidCreatedAt

						return
					}

					createdAt, err := time.Parse(time.RFC3339Nano, createdAtRFC)
					if err != nil {
						createdAt, err = time.Parse("2006-01-02T15:04:05.999999", createdAtRFC) // For some reason, Bsky sometimes seems to not specify the timezone
						if err != nil {
							errs <- errMessageInvalidCreatedAt

							return
						}
					}

					rawText, ok := message.Values["text"]
					if !ok {
						errs <- errMessageMissingText

						return
					}

					text, ok := rawText.(string)
					if !ok {
						errs <- errMessageInvalidText

						return
					}

					rawReply, ok := message.Values["reply"]
					if !ok {
						errs <- errMessageMissingReply

						return
					}

					replyValue, ok := rawReply.(string)
					if !ok {
						errs <- errMessageInvalidReply

						return
					}

					reply := replyValue == "true"

					rawLangs, ok := message.Values["langs"]
					if !ok {
						errs <- errMessageMissingLangs

						return
					}

					langsJoined, ok := rawLangs.(string)
					if !ok {
						errs <- errMessageInvalidLangs

						return
					}

					langs := strings.Split(langsJoined, ",")

					post, err := persister.CreatePost(
						ctx,
						did,
						rkey,
						createdAt,
						text,
						reply,
						langs,
					)
					if err != nil {
						log.Println("Could not insert post, skipping:", err)

						continue
					}

					if *verbose {
						log.Println("Created post", post)
					}

					if err := classify(post); err != nil {
						log.Println("Could not classify post, skipping:", err)

						continue
					}
				}
			}
		}
	}()

	go func() {
		for {
			streams, err := broker.XReadGroup(ctx, &redis.XReadGroupArgs{
				Group:    persisters.StreamPostLike,
				Consumer: uuid.NewString(),
				Streams:  []string{persisters.StreamPostLike, ">"},
				Block:    0,
				Count:    10,
			}).Result()
			if err != nil {
				log.Println("Could not subscribe to post like stream, skipping:", err)

				continue
			}

			for _, stream := range streams {
				for _, message := range stream.Messages {
					rawDid, ok := message.Values["did"]
					if !ok {
						errs <- errMessageMissingDID

						return
					}

					did, ok := rawDid.(string)
					if !ok {
						errs <- errMessageInvalidDID

						return
					}

					rawRkey, ok := message.Values["rkey"]
					if !ok {
						errs <- errMessageMissingRkey

						return
					}

					rkey, ok := rawRkey.(string)
					if !ok {
						errs <- errMessageInvalidRkey

						return
					}

					post, err := persister.LikePost(
						ctx,
						did,
						rkey,
					)
					if err != nil {
						log.Println("Could not like post, skipping:", err)

						continue
					}

					if *verbose {
						log.Println("Liked post", post)
					}

					if err := classify(post); err != nil {
						log.Println("Could not classify post, skipping:", err)

						continue
					}
				}
			}
		}
	}()

	for err := range errs {
		if err != nil {
			panic(err)
		}
	}
}
