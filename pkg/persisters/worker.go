package persisters

import (
	"database/sql"

	"github.com/pojntfx/atmosfeed/pkg/models"
	"github.com/redis/go-redis/v9"
)

type WorkerPersister struct {
	pgaddr  string
	queries *models.Queries
	db      *sql.DB

	broker *redis.Client
}

func NewWorkerPersister(pgaddr string, broker *redis.Client) *WorkerPersister {
	return &WorkerPersister{
		pgaddr: pgaddr,

		broker: broker,
	}
}

func (p *WorkerPersister) Init() error {
	var err error
	p.db, err = sql.Open("postgres", p.pgaddr)
	if err != nil {
		return err
	}

	p.queries = models.New(p.db)

	return nil
}
