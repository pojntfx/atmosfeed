package trending

import (
	"signature"
)

func Scale(ctx *signature.Context) (*signature.Context, error) {
	if ctx.Post.Likes >= 10 {
		ctx.Weight = ctx.Post.Likes
	} else {
		ctx.Weight = -1
	}

	return signature.Next(ctx)
}
