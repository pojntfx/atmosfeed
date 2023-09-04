package questions

import (
	"signature"
	"strings"
)

func Scale(ctx *signature.Context) (*signature.Context, error) {
	if strings.Contains(ctx.Post.Text, "?") {
		ctx.Weight = ctx.Post.Likes
	} else {
		ctx.Weight = -1
	}

	return signature.Next(ctx)
}
