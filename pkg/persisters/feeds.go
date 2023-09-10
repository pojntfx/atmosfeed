package persisters

import (
	"context"
	"path"
	"time"

	"github.com/pojntfx/atmosfeed/pkg/models"
)

func (p *ManagerPersister) UpsertFeed(
	ctx context.Context,
	did string,
	rkey string,
	classifier []byte,
) error {
	if err := p.queries.UpsertFeed(ctx, models.UpsertFeedParams{
		Did:        did,
		Rkey:       rkey,
		Classifier: classifier,
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

func (p *WorkerPersister) GetFeedClassifier(
	ctx context.Context,
	did string,
	rkey string,
) ([]byte, error) {
	return p.queries.GetFeedClassifier(ctx, models.GetFeedClassifierParams{
		Did:  did,
		Rkey: rkey,
	})
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
		FeedDid:   feedDid,
		FeedRkey:  feedRkey,
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
