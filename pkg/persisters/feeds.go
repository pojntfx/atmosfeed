package persisters

import (
	"context"
	"time"

	"github.com/pojntfx/atmosfeed/pkg/models"
)

func (p *Persister) UpsertFeed(
	ctx context.Context,
	did string,
	rkey string,
	classifier []byte,
) error {
	return p.queries.UpsertFeed(ctx, models.UpsertFeedParams{
		Did:        did,
		Rkey:       rkey,
		Classifier: classifier,
	})
}

func (p *Persister) GetFeeds(
	ctx context.Context,
) ([]models.Feed, error) {
	return p.queries.GetFeeds(ctx)
}

func (p *Persister) GetFeedClassifier(
	ctx context.Context,
	did string,
	rkey string,
) ([]byte, error) {
	return p.queries.GetFeedClassifier(ctx, models.GetFeedClassifierParams{
		Did:  did,
		Rkey: rkey,
	})
}

func (p *Persister) DeleteFeed(
	ctx context.Context,
	did string,
	rkey string,
) error {
	return p.queries.DeleteFeed(ctx, models.DeleteFeedParams{
		Did:  did,
		Rkey: rkey,
	})
}

func (p *Persister) CreateFeedPost(
	ctx context.Context,
	feedDid string,
	feedRkey string,
	postDid string,
	postRkey string,
) error {
	return p.queries.CreateFeedPost(ctx, models.CreateFeedPostParams{
		FeedDid:  feedDid,
		FeedRkey: feedRkey,
		PostDid:  postDid,
		PostRkey: postRkey,
	})
}

func (p *Persister) GetFeedPosts(
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

func (p *Persister) GetFeedPostsCursor(
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
