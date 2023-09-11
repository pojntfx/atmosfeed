package persisters

//go:generate sqlc -f ../../sqlc.yaml generate

import (
	"context"
	"database/sql"
	"net/url"
	"strings"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/pojntfx/atmosfeed/pkg/migrations"
	"github.com/pojntfx/atmosfeed/pkg/models"
	"github.com/pressly/goose/v3"
	"github.com/redis/go-redis/v9"
)

const (
	TopicFeedUpsert = "feed/upsert"
	TopicFeedDelete = "feed/delete"

	StreamPostInsert = "post/insert"
	StreamPostLike   = "post/like"

	errBusyGroup = "BUSYGROUP Consumer Group name already exists"
)

type ManagerPersister struct {
	pgaddr  string
	queries *models.Queries
	db      *sql.DB
	s3url   string

	broker *redis.Client
	blobs  *minio.Client
	bucket string
}

func NewManagerPersister(pgaddr string, broker *redis.Client, s3url string) *ManagerPersister {
	return &ManagerPersister{
		pgaddr: pgaddr,
		s3url:  s3url,

		broker: broker,
	}
}

func (p *ManagerPersister) Init(ctx context.Context) error {
	if _, err := p.broker.XGroupCreateMkStream(ctx, StreamPostInsert, StreamPostInsert, "$").Result(); err != nil && !strings.Contains(err.Error(), errBusyGroup) {
		return err
	}

	if _, err := p.broker.XGroupCreateMkStream(ctx, StreamPostLike, StreamPostLike, "$").Result(); err != nil && !strings.Contains(err.Error(), errBusyGroup) {
		return err
	}

	var err error
	p.db, err = sql.Open("postgres", p.pgaddr)
	if err != nil {
		return err
	}

	goose.SetBaseFS(migrations.FS)

	if err := goose.SetDialect("postgres"); err != nil {
		return err
	}

	if err := goose.Up(p.db, "."); err != nil {
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

	bucketExists, err := p.blobs.BucketExists(ctx, p.bucket)
	if err != nil {
		return err
	}

	if !bucketExists {
		if err := p.blobs.MakeBucket(ctx, p.bucket, minio.MakeBucketOptions{}); err != nil {
			return err
		}
	}

	return nil
}
