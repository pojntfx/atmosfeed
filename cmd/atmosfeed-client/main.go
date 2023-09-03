package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/bluesky-social/indigo/api/atproto"
	"github.com/bluesky-social/indigo/api/bsky"
	"github.com/bluesky-social/indigo/lex/util"
	"github.com/bluesky-social/indigo/xrpc"
	"github.com/pojntfx/atmosfeed/pkg/persisters"
)

const (
	lexiconFeedGenerator = "app.bsky.feed.generator"
)

func main() {
	pdsURL := flag.String("pds-url", "https://bsky.social", "PDS URL")
	username := flag.String("username", "example.bsky.social", "Bluesky username (ignored if `--publish` is not provided)")
	password := flag.String("password", "", "Bluesky password, preferably an app password (get one from https://bsky.app/settings/app-passwords, ignored if `--publish` is not provided)")

	postgresURL := flag.String("postgres-url", "postgresql://postgres@localhost:5432/atmosfeed?sslmode=disable", "PostgreSQL URL")

	feedGeneratorDID := flag.String("feed-generator-did", "did:web:atmosfeed-feeds.serveo.net", "DID of the feed generator (typically the hostname of the publicly reachable URL) (ignored for `--delete`)")

	feedRkey := flag.String("feed-rkey", "trending", "Machine-readable key for the feed")
	feedName := flag.String("feed-name", "Atmosfeed Trending", "Human-readable name for the feed (ignored for `--delete`)")
	feedDescription := flag.String("feed-description", "An example trending feed for Atmosfeed", "Description for the feed (ignored for `--delete`)")
	feedClassifier := flag.String("feed-classifier", "out/local-trending-latest.scale", "Path to the feed classifier to upload (ignored for `--delete` or if `--publish` is not provided)")

	delete := flag.Bool("delete", false, "Whether to delete instead of upsert a feed/classifier")
	publish := flag.Bool("publish", true, "Whether to publish/unpublish the feed to/from Blueksy")

	flag.Parse()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	auth := &xrpc.AuthInfo{}

	client := &xrpc.Client{
		Client: http.DefaultClient,
		Host:   *pdsURL,
		Auth:   auth,
	}

	session, err := atproto.ServerCreateSession(ctx, client, &atproto.ServerCreateSession_Input{
		Identifier: *username,
		Password:   *password,
	})
	if err != nil {
		panic(err)
	}

	auth.AccessJwt = session.AccessJwt
	auth.RefreshJwt = session.RefreshJwt
	auth.Handle = session.Handle
	auth.Did = session.Did

	log.Println("Connected to PDS", *pdsURL)

	var persister *persisters.Persister
	if *publish {
		persister = persisters.NewPersister(*postgresURL)

		if err := persister.Init(false); err != nil {
			panic(err)
		}

		log.Println("Connected to PostgreSQL")
	}

	rec := &util.LexiconTypeDecoder{
		Val: &bsky.FeedGenerator{
			CreatedAt:   time.Now().Format(time.RFC3339),
			Description: feedDescription,
			Did:         *feedGeneratorDID,
			DisplayName: *feedName,
		},
	}

	if *delete {
		if err := atproto.RepoDeleteRecord(ctx, client, &atproto.RepoDeleteRecord_Input{
			Collection: lexiconFeedGenerator,
			Repo:       auth.Did,
			Rkey:       *feedRkey,
		}); err != nil {
			panic(err)
		}

		if *publish {
			if err := persister.DeleteFeed(ctx, *feedRkey); err != nil {
				panic(err)
			}
		}
	} else {
		if *publish {
			b, err := os.ReadFile(*feedClassifier)
			if err != nil {
				panic(err)
			}

			if err := persister.UpsertFeed(ctx, *feedRkey, b); err != nil {
				panic(err)
			}
		}

		ex, err := atproto.RepoGetRecord(ctx, client, "", lexiconFeedGenerator, auth.Did, *feedRkey)
		if err == nil {
			if _, err := atproto.RepoPutRecord(ctx, client, &atproto.RepoPutRecord_Input{
				Collection: lexiconFeedGenerator,
				Repo:       auth.Did,
				Rkey:       *feedRkey,
				Record:     rec,
				SwapRecord: ex.Cid,
			}); err != nil {
				panic(err)
			}
		} else {
			if _, err := atproto.RepoCreateRecord(ctx, client, &atproto.RepoCreateRecord_Input{
				Collection: lexiconFeedGenerator,
				Repo:       auth.Did,
				Rkey:       feedRkey,
				Record:     rec,
			}); err != nil {
				panic(err)
			}
		}
	}
}
