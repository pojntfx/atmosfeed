package trending

import (
	"signature"
)

func Scale(ctx *signature.Context) (*signature.Context, error) {
	if ctx.Post.Likes > 10 {
		ctx.Include = true
	}

	return signature.Next(ctx)
}
