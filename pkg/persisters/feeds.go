package persisters

import (
	"context"
	"io"
	"path"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/pojntfx/atmosfeed/pkg/models"
)

func (p *ManagerPersister) UpsertFeed(
	ctx context.Context,
	did string,
	rkey string,
	pinnedDID string,
	pinnedRkey string,
	classifier io.Reader,
) error {
	if _, err := p.blobs.PutObject(
		ctx,
		p.bucket,
		path.Join(did, rkey),
		classifier,
		-1,
		minio.PutObjectOptions{},
	); err != nil {
		return err
	}

	if err := p.queries.UpsertFeed(ctx, models.UpsertFeedParams{
		Did:        did,
		Rkey:       rkey,
		PinnedDid:  pinnedDID,
		PinnedRkey: pinnedRkey,
	}); err != nil {
		return err
	}

	if _, err := p.broker.Publish(ctx, TopicFeedUpsert, path.Join(did, rkey)).Result(); err != nil {
		return err
	}

	return nil
}

func (p *WorkerPersister) GetFeeds(
	ctx context.Context,
) ([]models.Feed, error) {
	return p.queries.GetFeeds(ctx)
}

func (p *ManagerPersister) GetFeedsForDid(
	ctx context.Context,
	did string,
) ([]models.Feed, error) {
	return p.queries.GetFeedsForDid(ctx, did)
}

func (p *ManagerPersister) GetFeedClassifier(
	ctx context.Context,
	did string,
	rkey string,
) (io.Reader, error) {
	return p.blobs.GetObject(
		ctx,
		p.bucket,
		path.Join(did, rkey),
		minio.GetObjectOptions{},
	)
}

func (p *WorkerPersister) GetFeedClassifier(
	ctx context.Context,
	did string,
	rkey string,
) (io.Reader, error) {
	return p.blobs.GetObject(
		ctx,
		p.bucket,
		path.Join(did, rkey),
		minio.GetObjectOptions{},
	)
}

func (p *ManagerPersister) DeleteFeed(
	ctx context.Context,
	did string,
	rkey string,
) error {
	if err := p.queries.DeleteFeed(ctx, models.DeleteFeedParams{
		Did:  did,
		Rkey: rkey,
	}); err != nil {
		return err
	}

	if err := p.blobs.RemoveObject(ctx, p.bucket, path.Join(did, rkey), minio.RemoveObjectOptions{}); err != nil {
		return err
	}

	if _, err := p.broker.Publish(ctx, TopicFeedDelete, path.Join(did, rkey)).Result(); err != nil {
		return err
	}

	return nil
}

func (p *WorkerPersister) UpsertFeedPost(
	ctx context.Context,
	feedDid string,
	feedRkey string,
	postDid string,
	postRkey string,
	weight int32,
) error {
	return p.queries.UpsertFeedPost(ctx, models.UpsertFeedPostParams{
		FeedDid:  feedDid,
		FeedRkey: feedRkey,
		PostDid:  postDid,
		PostRkey: postRkey,
		Weight:   weight,
	})
}

func (p *ManagerPersister) GetFeedPosts(
	ctx context.Context,
	feedDid string,
	feedRkey string,
	ttl time.Time,
	limit int32,
) ([]models.GetFeedPostsRow, error) {
	return p.queries.GetFeedPosts(ctx, models.GetFeedPostsParams{
		Did:       feedDid,
		Rkey:      feedRkey,
		CreatedAt: ttl,
		Limit:     limit,
	})
}

func (p *ManagerPersister) GetFeedPostsCursor(
	ctx context.Context,
	feedDid string,
	feedRkey string,
	ttl time.Time,
	limit int32,
	postDid string,
	postRkey string,
) ([]models.GetFeedPostsCursorRow, error) {
	return p.queries.GetFeedPostsCursor(ctx, models.GetFeedPostsCursorParams{
		FeedDid:   feedDid,
		FeedRkey:  feedRkey,
		CreatedAt: ttl,
		Limit:     limit,
		Did:       postDid,
		Rkey:      postRkey,
	})
}

func (p *ManagerPersister) GetFeedPostsForDid(
	ctx context.Context,
	did string,
) ([]models.FeedPost, error) {
	return p.queries.GetFeedPostsForDid(ctx, did)
}

func (p *ManagerPersister) DeleteFeedPostsForDid(
	ctx context.Context,
	did string,
) error {
	return p.queries.DeleteFeedPostsForDid(ctx, did)
}
