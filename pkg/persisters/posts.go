package persisters

import (
	"context"
	"time"

	"github.com/pojntfx/atmosfeed/pkg/models"
)

func (p *WorkerPersister) CreatePost(
	ctx context.Context,
	did string,
	rkey string,
	createdAt time.Time,
	text string,
	reply bool,
	langs []string,
) (models.Post, error) {
	return p.queries.CreatePost(ctx, models.CreatePostParams{
		Did:       did,
		Rkey:      rkey,
		CreatedAt: createdAt,
		Text:      text,
		Reply:     reply,
		Langs:     langs,
	})
}

func (p *WorkerPersister) LikePost(
	ctx context.Context,
	did string,
	rkey string,
) (models.Post, error) {
	return p.queries.LikePost(ctx, models.LikePostParams{
		Did:  did,
		Rkey: rkey,
	})
}
