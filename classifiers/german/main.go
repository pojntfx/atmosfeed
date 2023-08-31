package german

import (
	"signature"

	"golang.org/x/exp/slices"
)

func Scale(ctx *signature.Context) (*signature.Context, error) {
	if slices.Contains(ctx.Post.Langs, "de") {
		ctx.Include = true
	}

	return signature.Next(ctx)
}
