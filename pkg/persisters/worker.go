package persisters

import (
	"database/sql"
	"net/url"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/pojntfx/atmosfeed/pkg/models"
	"github.com/redis/go-redis/v9"
)

type WorkerPersister struct {
	pgaddr  string
	queries *models.Queries
	db      *sql.DB
	s3url   string

	broker *redis.Client
	blobs  *minio.Client
	bucket string
}

func NewWorkerPersister(pgaddr string, broker *redis.Client, s3url string) *WorkerPersister {
	return &WorkerPersister{
		pgaddr: pgaddr,
		s3url:  s3url,

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

	u, err := url.Parse(p.s3url)
	if err != nil {
		return err
	}

	user := u.User
	pw, _ := user.Password()

	p.blobs, err = minio.New(u.Host, &minio.Options{
		Creds:  credentials.NewStaticV4(user.Username(), pw, ""),
		Secure: u.Scheme == "https",
	})
	if err != nil {
		return err
	}

	p.bucket = u.Query().Get("bucket")

	return nil
}
