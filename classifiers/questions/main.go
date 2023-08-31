package questions

import (
	"signature"
	"strings"
)

func Scale(ctx *signature.Context) (*signature.Context, error) {
	if strings.Contains(ctx.Post.Text, "?") {
		ctx.Include = true
	}

	return signature.Next(ctx)
}
