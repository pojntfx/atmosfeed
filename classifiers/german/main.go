package german

import (
	"signature"
)

func Scale(ctx *signature.Context) (*signature.Context, error) {
	if len(ctx.Post.Langs) == 1 && ctx.Post.Langs[0] == "de" {
		ctx.Weight = ctx.Post.Likes
	} else {
		ctx.Weight = -1
	}

	return signature.Next(ctx)
}
