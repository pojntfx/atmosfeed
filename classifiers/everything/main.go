package everything

import (
	"signature"
)

func Scale(ctx *signature.Context) (*signature.Context, error) {
	ctx.Include = true

	return signature.Next(ctx)
}
