// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.20.0

package models

import (
	"time"
)

type Post struct {
	ID        int32
	CreatedAt time.Time
	Did       string
	Rkey      string
	Text      string
	Reply     bool
	Langs     []string
	Likes     int32
}
