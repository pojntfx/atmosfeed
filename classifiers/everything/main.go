package everything

import (
	"signature"
)

func Scale(ctx *signature.Context) (*signature.Context, error) {
	ctx.Weight = ctx.Post.CreatedAt

	return signature.Next(ctx)
}
