package main

import (
	"context"
	"flag"
	"log"
	"os"

	"github.com/pojntfx/atmosfeed/pkg/persisters"
)

func main() {
	postgresURL := flag.String("postgres-url", "postgresql://postgres@localhost:5432/atmosfeed?sslmode=disable", "PostgreSQL URL")

	feedRkey := flag.String("feed-rkey", "trending", "Machine-readable key for the feed")
	feedClassifier := flag.String("feed-classifier", "out/local-trending-latest.scale", "Path to the feed classifier to upload (ignored for `--delete`)")

	delete := flag.Bool("delete", false, "Whether to delete instead of upsert a feed/classifier")

	flag.Parse()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	persister := persisters.NewPersister(*postgresURL)

	if err := persister.Init(); err != nil {
		panic(err)
	}

	log.Println("Connected to PostgreSQL")

	if *delete {
		if err := persister.DeleteFeed(ctx, *feedRkey); err != nil {
			panic(err)
		}
	} else {
		b, err := os.ReadFile(*feedClassifier)
		if err != nil {
			panic(err)
		}

		if err := persister.UpsertFeed(ctx, *feedRkey, b); err != nil {
			panic(err)
		}
	}
}
