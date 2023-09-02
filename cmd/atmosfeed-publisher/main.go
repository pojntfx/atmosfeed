package main

import (
	"context"
	"flag"
	"net/http"
	"time"

	"github.com/bluesky-social/indigo/api/atproto"
	"github.com/bluesky-social/indigo/api/bsky"
	"github.com/bluesky-social/indigo/lex/util"
	"github.com/bluesky-social/indigo/xrpc"
)

const (
	lexiconFeedGenerator = "app.bsky.feed.generator"
)

func main() {
	pdsURL := flag.String("pds-url", "https://bsky.social", "PDS URL")
	username := flag.String("username", "example.bsky.social", "Bluesky username to login with")
	password := flag.String("password", "", "Bluesky password to login with, preferably an app password (get one from https://bsky.app/settings/app-passwords)")

	feedGeneratorDID := flag.String("feed-generator-did", "did:web:atmosfeed-feeds.serveo.net", "DID of the feed generator (typically the hostname of the publicly reachable URL)")

	feedRkey := flag.String("feed-rkey", "trending", "Machine-readable key for the feed")
	feedName := flag.String("feed-name", "Atmosfeed Trending", "Human-readable name for the feed")
	feedDescription := flag.String("feed-description", "An example trending feed for Atmosfeed", "Description for the feed")

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

	rec := &util.LexiconTypeDecoder{&bsky.FeedGenerator{
		CreatedAt:   time.Now().Format(time.RFC3339),
		Description: feedDescription,
		Did:         *feedGeneratorDID,
		DisplayName: *feedName,
	}}

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
