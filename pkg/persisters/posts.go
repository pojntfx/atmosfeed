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

func (p *ManagerPersister) DeletePost(
	ctx context.Context,
	did string,
	rkey string,
) error {
	return p.queries.DeletePost(ctx, models.DeletePostParams{
		Did:  did,
		Rkey: rkey,
	})
}

func (p *ManagerPersister) DeleteAllPosts(
	ctx context.Context,
) error {
	return p.queries.DeleteAllPosts(ctx)
}

func (p *ManagerPersister) GetPostsForDid(
	ctx context.Context,
	did string,
) ([]models.Post, error) {
	return p.queries.GetPostsForDid(ctx, did)
}

func (p *ManagerPersister) DeletePostsForDid(
	ctx context.Context,
	did string,
) error {
	return p.queries.DeletePostsForDid(ctx, did)
}
