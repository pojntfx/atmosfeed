package persisters

import (
	"context"
	"time"

	"github.com/pojntfx/atmosfeed/pkg/models"
)

func (p *Persister) CreatePost(
	ctx context.Context,
	did string,
	rkey string,
	createdAt time.Time,
	text string,
	reply bool,
	langs []string,
) error {
	return p.queries.CreatePost(ctx, models.CreatePostParams{
		Did:       did,
		Rkey:      rkey,
		CreatedAt: createdAt,
		Text:      text,
		Reply:     reply,
		Langs:     langs,
	})
}

func (p *Persister) LikePost(
	ctx context.Context,
	did string,
	rkey string,
) (models.Post, error) {
	return p.queries.LikePost(ctx, models.LikePostParams{
		Did:  did,
		Rkey: rkey,
	})
}
