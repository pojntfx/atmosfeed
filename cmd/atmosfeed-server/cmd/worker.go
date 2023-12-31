package cmd

import (
	"context"
	"database/sql"
	"errors"
	"io"
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
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	classifierTimeoutFlag = "classifier-timeout"
	workingDirectoryFlag  = "working-directory"

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

var workerCmd = &cobra.Command{
	Use:     "worker",
	Aliases: []string{"w"},
	Short:   "Start an Atmosfeed worker",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := viper.BindPFlags(cmd.PersistentFlags()); err != nil {
			return err
		}

		if err := os.RemoveAll(filepath.Join(viper.GetString(workingDirectoryFlag), classifiersPath)); err != nil {
			panic(err)
		}

		options, err := redis.ParseURL(viper.GetString(redisURLFlag))
		if err != nil {
			panic(err)
		}

		broker := redis.NewClient(options)
		defer broker.Close()

		log.Println("Connected to Redis")

		persister := persisters.NewWorkerPersister(viper.GetString(postgresURLFlag), broker, viper.GetString(s3URLFlag))

		if err := persister.Init(); err != nil {
			panic(err)
		}

		log.Println("Connected to PostgreSQL")

		var classifierLock sync.Mutex
		classifiers := map[string]*scale.Instance[*signature.Signature]{}

		fetchClassifier := func(did, rkey string) error {
			classifierSource, err := persister.GetFeedClassifier(cmd.Context(), did, rkey)
			if err != nil {
				return err
			}

			classifierPath := filepath.Join(viper.GetString(workingDirectoryFlag), classifiersPath, did, rkey)
			if err := os.MkdirAll(filepath.Dir(classifierPath), os.ModePerm); err != nil {
				return err
			}

			classifierLock.Lock()
			defer classifierLock.Unlock()

			f, err := os.OpenFile(classifierPath, os.O_RDWR|os.O_CREATE, os.ModePerm)
			if err != nil {
				return err
			}
			defer f.Close()

			if _, err := io.Copy(f, classifierSource); err != nil {
				return err
			}

			fn, err := scalefunc.Read(classifierPath)
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

			classifiers[path.Join(did, rkey)] = classifier

			return nil
		}

		errs := make(chan error)

		go func() {
			streams := broker.Subscribe(cmd.Context(), persisters.TopicFeedUpsert)
			defer streams.Close()

			messages := streams.Channel()
			for message := range messages {
				did, rkey := path.Dir(message.Payload), path.Base(message.Payload)

				if err := fetchClassifier(did, rkey); err != nil {
					log.Println("Could not fetch classifier, skipping:", err)

					continue
				}

				if viper.GetBool(verboseFlag) {
					log.Println("Upserted classifier for feed", did, rkey)
				}
			}
		}()

		go func() {
			streams := broker.Subscribe(cmd.Context(), persisters.TopicFeedDelete)
			defer streams.Close()

			messages := streams.Channel()
			for message := range messages {
				did, rkey := path.Dir(message.Payload), path.Base(message.Payload)

				func() {
					classifierLock.Lock()
					defer classifierLock.Unlock()

					if err := os.RemoveAll(filepath.Join(viper.GetString(workingDirectoryFlag), classifiersPath, did, rkey)); err != nil {
						log.Println("Could not remove classifier from disk, skipping:", err)

						return
					}

					delete(classifiers, path.Join(did, rkey))

					if viper.GetBool(verboseFlag) {
						log.Println("Deleted feed", did, rkey)
					}
				}()
			}
		}()

		classifierSources, err := persister.GetFeeds(cmd.Context())
		if err != nil {
			panic(err)
		}

		for _, classifierSource := range classifierSources {
			did, rkey := classifierSource.Did, classifierSource.Rkey

			if err := fetchClassifier(did, rkey); err != nil {
				log.Println("Could not fetch classifier, skipping:", err)

				continue
			}

			if viper.GetBool(verboseFlag) {
				log.Println("Fetched classifier for feed", did, rkey)
			}
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

					ctx, cancel := context.WithTimeout(context.Background(), viper.GetDuration(classifierTimeoutFlag))
					defer cancel()

					if err := classifier.Run(ctx, s); err != nil {
						errs <- err

						return
					}

					if s.Context.Weight >= 0 {
						if err := persister.UpsertFeedPost(cmd.Context(), feedDid, feedRkey, p.Did, p.Rkey, int32(s.Context.Weight)); err != nil {
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
				streams, err := broker.XReadGroup(cmd.Context(), &redis.XReadGroupArgs{
					Group:    persisters.StreamPostInsert,
					Consumer: uuid.NewString(),
					Streams:  []string{persisters.StreamPostInsert, ">"},
					Block:    0,
					Count:    10,
				}).Result()
				if err != nil {
					errs <- err

					return
				}

				for _, stream := range streams {
					for _, message := range stream.Messages {
						rawDid, ok := message.Values["did"]
						if !ok {
							log.Println(errMessageMissingDID)

							continue
						}

						did, ok := rawDid.(string)
						if !ok {
							log.Println(errMessageInvalidDID)

							continue
						}

						rawRkey, ok := message.Values["rkey"]
						if !ok {
							log.Println(errMessageMissingRkey)

							continue
						}

						rkey, ok := rawRkey.(string)
						if !ok {
							log.Println(errMessageInvalidRkey)

							continue
						}

						rawCreatedAt, ok := message.Values["createdAt"]
						if !ok {
							log.Println(errMessageMissingCreatedAt)

							continue
						}

						createdAtRFC, ok := rawCreatedAt.(string)
						if !ok {
							log.Println(errMessageInvalidCreatedAt)

							continue
						}

						createdAt, err := time.Parse(time.RFC3339Nano, createdAtRFC)
						if err != nil {
							createdAt, err = time.Parse("2006-01-02T15:04:05.999999", createdAtRFC) // For some reason, Bsky sometimes seems to not specify the timezone
							if err != nil {
								log.Println(errMessageInvalidCreatedAt)

								continue
							}
						}

						rawText, ok := message.Values["text"]
						if !ok {
							log.Println(errMessageMissingText)

							continue
						}

						text, ok := rawText.(string)
						if !ok {
							log.Println(errMessageInvalidText)

							continue
						}

						rawReply, ok := message.Values["reply"]
						if !ok {
							log.Println(errMessageMissingReply)

							continue
						}

						replyValue, ok := rawReply.(string)
						if !ok {
							log.Println(errMessageInvalidReply)

							continue
						}

						reply := replyValue == "true"

						rawLangs, ok := message.Values["langs"]
						if !ok {
							log.Println(errMessageMissingLangs)

							continue
						}

						langsJoined, ok := rawLangs.(string)
						if !ok {
							log.Println(errMessageInvalidLangs)

							continue
						}

						langs := strings.Split(langsJoined, ",")

						post, err := persister.CreatePost(
							cmd.Context(),
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

						if viper.GetBool(verboseFlag) {
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
				streams, err := broker.XReadGroup(cmd.Context(), &redis.XReadGroupArgs{
					Group:    persisters.StreamPostLike,
					Consumer: uuid.NewString(),
					Streams:  []string{persisters.StreamPostLike, ">"},
					Block:    0,
					Count:    10,
				}).Result()
				if err != nil {
					errs <- err

					return
				}

				for _, stream := range streams {
					for _, message := range stream.Messages {
						rawDid, ok := message.Values["did"]
						if !ok {
							log.Println(errMessageMissingDID)

							continue
						}

						did, ok := rawDid.(string)
						if !ok {
							log.Println(errMessageInvalidDID)

							continue
						}

						rawRkey, ok := message.Values["rkey"]
						if !ok {
							log.Println(errMessageMissingRkey)

							continue
						}

						rkey, ok := rawRkey.(string)
						if !ok {
							log.Println(errMessageInvalidRkey)

							continue
						}

						post, err := persister.LikePost(
							cmd.Context(),
							did,
							rkey,
						)
						if err != nil && !errors.Is(err, sql.ErrNoRows) {
							log.Println("Could not like post, skipping:", err)

							continue
						}

						if viper.GetBool(verboseFlag) {
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
			if err == nil {
				return nil
			}

			return err
		}

		return nil
	},
}

func init() {
	home, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}

	workerCmd.PersistentFlags().Duration(classifierTimeoutFlag, time.Second, "Amount of time after which to stop a classifier Scale function from running")
	workerCmd.PersistentFlags().String(workingDirectoryFlag, filepath.Join(home, ".local", "share", "atmosfeed", "var", "lib", "atmosfeed"), "Working directory to use")

	viper.AutomaticEnv()

	rootCmd.AddCommand(workerCmd)
}
