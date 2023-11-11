// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.23.0

package models

import (
	"time"
)

type Feed struct {
	Did        string
	Rkey       string
	PinnedDid  string
	PinnedRkey string
}

type FeedPost struct {
	FeedDid  string
	FeedRkey string
	PostDid  string
	PostRkey string
	Weight   int32
}

type Post struct {
	Did       string
	Rkey      string
	CreatedAt time.Time
	Text      string
	Reply     bool
	Langs     []string
	Likes     int32
}
