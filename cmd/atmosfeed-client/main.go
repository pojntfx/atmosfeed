package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
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
	username := flag.String("username", "example.bsky.social", "Bluesky username (ignored if `--publish` is not provided)")
	password := flag.String("password", "", "Bluesky password, preferably an app password (get one from https://bsky.app/settings/app-passwords, ignored if `--publish` is not provided)")

	atmosfeedURL := flag.String("atmosfeed-url", "http://localhost:1337", "Atmosfeed server URL")

	feedGeneratorDID := flag.String("feed-generator-did", "did:web:atmosfeed-feeds.serveo.net", "DID of the feed generator (typically the hostname of the publicly reachable URL) (ignored for `--delete`)")

	feedRkey := flag.String("feed-rkey", "trending", "Machine-readable key for the feed")
	feedName := flag.String("feed-name", "Atmosfeed Trending", "Human-readable name for the feed (ignored for `--delete`)")
	feedDescription := flag.String("feed-description", "An example trending feed for Atmosfeed", "Description for the feed (ignored for `--delete`)")
	feedClassifier := flag.String("feed-classifier", "out/local-trending-latest.scale", "Path to the feed classifier to upload (ignored for `--delete` or if `--publish` is not provided)")

	delete := flag.Bool("delete", false, "Whether to delete instead of upsert a feed/classifier")
	publish := flag.Bool("publish", false, "Whether to publish/unpublish the feed to/from Blueksy")
	list := flag.Bool("list", false, "Whether to just list available feeds in Atmosfeed and then exit")

	flag.Parse()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	u, err := url.Parse(*atmosfeedURL)
	if err != nil {
		panic(err)
	}

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

	rec := &util.LexiconTypeDecoder{
		Val: &bsky.FeedGenerator{
			CreatedAt:   time.Now().Format(time.RFC3339),
			Description: feedDescription,
			Did:         *feedGeneratorDID,
			DisplayName: *feedName,
		},
	}

	if *list {
		u := u.JoinPath("admin", "feeds")

		q := u.Query()
		q.Add("rkey", *feedRkey)
		u.RawQuery = q.Encode()

		req, err := http.NewRequest(http.MethodGet, u.String(), nil)
		if err != nil {
			panic(err)
		}

		req.Header.Set("Authorization", "Bearer "+auth.AccessJwt)

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			panic(err)
		}
		defer resp.Body.Close()

		feeds := []string{}
		if err := json.NewDecoder(resp.Body).Decode(&feeds); err != nil {
			panic(err)
		}

		if resp.StatusCode != http.StatusOK {
			panic(resp.Status)
		}

		fmt.Println(strings.Join(feeds, ","))

		return
	}

	if *delete {
		if *publish {
			if err := atproto.RepoDeleteRecord(ctx, client, &atproto.RepoDeleteRecord_Input{
				Collection: lexiconFeedGenerator,
				Repo:       auth.Did,
				Rkey:       *feedRkey,
			}); err != nil {
				panic(err)
			}
		}

		u := u.JoinPath("admin", "feeds")

		q := u.Query()
		q.Add("rkey", *feedRkey)
		u.RawQuery = q.Encode()

		req, err := http.NewRequest(http.MethodDelete, u.String(), nil)
		if err != nil {
			panic(err)
		}

		req.Header.Set("Authorization", "Bearer "+auth.AccessJwt)

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			panic(err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			panic(resp.Status)
		}
	} else {
		f, err := os.Open(*feedClassifier)
		if err != nil {
			panic(err)
		}

		u := u.JoinPath("admin", "feeds")

		q := u.Query()
		q.Add("rkey", *feedRkey)
		u.RawQuery = q.Encode()

		req, err := http.NewRequest(http.MethodPut, u.String(), f)
		if err != nil {
			panic(err)
		}

		req.Header.Set("Authorization", "Bearer "+auth.AccessJwt)

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			panic(err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			panic(resp.Status)
		}

		if *publish {
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
}
