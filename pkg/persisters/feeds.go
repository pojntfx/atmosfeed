package persisters

import (
	"context"
	"time"

	"github.com/pojntfx/atmosfeed/pkg/models"
)

func (p *Persister) UpsertFeed(
	ctx context.Context,
	name string,
	classifier []byte,
) error {
	return p.queries.UpsertFeed(ctx, models.UpsertFeedParams{
		Name:       name,
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
	name string,
) ([]byte, error) {
	return p.queries.GetFeedClassifier(ctx, name)
}

func (p *Persister) DeleteFeed(
	ctx context.Context,
	name string,
) error {
	return p.queries.DeleteFeed(ctx, name)
}

func (p *Persister) CreateFeedPost(
	ctx context.Context,
	feedName string,
	postDid string,
	postKkey string,
) error {
	return p.queries.CreateFeedPost(ctx, models.CreateFeedPostParams{
		FeedName: feedName,
		PostDid:  postDid,
		PostRkey: postKkey,
	})
}

func (p *Persister) GetFeedPosts(
	ctx context.Context,
	feedName string,
	ttl time.Time,
	limit int32,
) ([]models.GetFeedPostsRow, error) {
	return p.queries.GetFeedPosts(ctx, models.GetFeedPostsParams{
		FeedName:  feedName,
		CreatedAt: ttl,
		Limit:     limit,
	})
}
