package persisters

import (
	"context"

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
