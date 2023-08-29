package persisters

import (
	"context"
	"time"

	"github.com/pojntfx/atmosfeed/pkg/models"
)

func (p *Persister) CreatePost(
	ctx context.Context,
	createdAt time.Time,
	did string,
	rkey string,
	text string,
	reply bool,
	langs []string,
) (int32, error) {
	return p.queries.CreatePost(ctx, models.CreatePostParams{
		CreatedAt: createdAt,
		Did:       did,
		Rkey:      rkey,
		Text:      text,
		Reply:     reply,
		Langs:     langs,
	})
}
