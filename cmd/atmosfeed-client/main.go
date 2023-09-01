package main

import (
	"context"
	"flag"
	"log"
	"os"

	"github.com/pojntfx/atmosfeed/pkg/persisters"
)

func main() {
	postgresURL := flag.String("postgres-url", "postgresql://postgres@localhost:5432/atmosfeed?sslmode=disable", "PostgreSQL URL (can also be set using `POSTGRES_URL` env variable)")
	name := flag.String("name", "trending", "Name of the feed")
	classifier := flag.String("classifier", "out/local-trending-latest.scale", "Path to the classifier to upload (ignored for `--delete`)")
	delete := flag.Bool("delete", false, "Whether to delete instead of upsert a classifier")

	flag.Parse()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if v := os.Getenv("POSTGRES_URL"); v != "" {
		log.Println("Using database address from POSTGRES_URL env variable")

		*postgresURL = v
	}

	p := persisters.NewPersister(*postgresURL)

	if err := p.Init(); err != nil {
		panic(err)
	}

	log.Println("Connected to PostgreSQL")

	if *delete {
		if err := p.DeleteFeed(ctx, *name); err != nil {
			panic(err)
		}
	} else {
		b, err := os.ReadFile(*classifier)
		if err != nil {
			panic(err)
		}

		if err := p.UpsertFeed(ctx, *name, b); err != nil {
			panic(err)
		}
	}
}
