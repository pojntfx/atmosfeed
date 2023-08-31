package alwaystrue

import (
	"signature"
)

func Scale(ctx *signature.Context) (*signature.Context, error) {
	ctx.MyString = "changed"

	return signature.Next(ctx)
}
