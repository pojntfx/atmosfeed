package persisters

//go:generate sqlc -f ../../sqlc.yaml generate

import (
	"context"
	"database/sql"
	"strings"

	"github.com/pojntfx/atmosfeed/pkg/migrations"
	"github.com/pojntfx/atmosfeed/pkg/models"
	"github.com/pressly/goose/v3"
	"github.com/redis/go-redis/v9"
)

const (
	StreamFeedUpsert = "feed/upsert"
	StreamFeedDelete = "feed/delete"

	errBusyGroup = "BUSYGROUP Consumer Group name already exists"
)

type Persister struct {
	pgaddr  string
	queries *models.Queries
	db      *sql.DB

	broker *redis.Client
}

func NewPersister(pgaddr string, broker *redis.Client) *Persister {
	return &Persister{
		pgaddr: pgaddr,

		broker: broker,
	}
}

func (p *Persister) Init(ctx context.Context, migrate bool) error {
	if _, err := p.broker.XGroupCreateMkStream(ctx, StreamFeedUpsert, StreamFeedUpsert, "$").Result(); err != nil && !strings.Contains(err.Error(), errBusyGroup) {
		return err
	}

	if _, err := p.broker.XGroupCreateMkStream(ctx, StreamFeedDelete, StreamFeedDelete, "$").Result(); err != nil && !strings.Contains(err.Error(), errBusyGroup) {
		return err
	}

	var err error
	p.db, err = sql.Open("postgres", p.pgaddr)
	if err != nil {
		return err
	}

	if migrate {
		goose.SetBaseFS(migrations.FS)

		if err := goose.SetDialect("postgres"); err != nil {
			return err
		}

		if err := goose.Up(p.db, "."); err != nil {
			return err
		}
	}

	p.queries = models.New(p.db)

	return nil
}
