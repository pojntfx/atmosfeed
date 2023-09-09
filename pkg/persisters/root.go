package persisters

//go:generate sqlc -f ../../sqlc.yaml generate

import (
	"database/sql"

	"github.com/pojntfx/atmosfeed/pkg/migrations"
	"github.com/pojntfx/atmosfeed/pkg/models"
	"github.com/pressly/goose/v3"
	"github.com/redis/go-redis/v9"
)

const (
	StreamFeedUpsert = "feed/upsert"
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

func (p *Persister) Init(migrate bool) error {
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
