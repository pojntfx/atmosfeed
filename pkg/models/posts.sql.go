// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.22.0
// source: posts.sql

package models

import (
	"context"
	"time"

	"github.com/lib/pq"
)

const createPost = `-- name: CreatePost :one
insert into posts (
        did,
        rkey,
        created_at,
        text,
        reply,
        langs,
        likes
    )
values ($1, $2, $3, $4, $5, $6, 0)
returning did, rkey, created_at, text, reply, langs, likes
`

type CreatePostParams struct {
	Did       string
	Rkey      string
	CreatedAt time.Time
	Text      string
	Reply     bool
	Langs     []string
}

func (q *Queries) CreatePost(ctx context.Context, arg CreatePostParams) (Post, error) {
	row := q.db.QueryRowContext(ctx, createPost,
		arg.Did,
		arg.Rkey,
		arg.CreatedAt,
		arg.Text,
		arg.Reply,
		pq.Array(arg.Langs),
	)
	var i Post
	err := row.Scan(
		&i.Did,
		&i.Rkey,
		&i.CreatedAt,
		&i.Text,
		&i.Reply,
		pq.Array(&i.Langs),
		&i.Likes,
	)
	return i, err
}

const deleteAllPosts = `-- name: DeleteAllPosts :exec
delete from posts
`

func (q *Queries) DeleteAllPosts(ctx context.Context) error {
	_, err := q.db.ExecContext(ctx, deleteAllPosts)
	return err
}

const deletePost = `-- name: DeletePost :exec
delete from posts
where did = $1
    and rkey = $2
`

type DeletePostParams struct {
	Did  string
	Rkey string
}

func (q *Queries) DeletePost(ctx context.Context, arg DeletePostParams) error {
	_, err := q.db.ExecContext(ctx, deletePost, arg.Did, arg.Rkey)
	return err
}

const deletePostsForDid = `-- name: DeletePostsForDid :exec
delete from posts
where did = $1
`

func (q *Queries) DeletePostsForDid(ctx context.Context, did string) error {
	_, err := q.db.ExecContext(ctx, deletePostsForDid, did)
	return err
}

const getPostsForDid = `-- name: GetPostsForDid :many
select did, rkey, created_at, text, reply, langs, likes
from posts
where did = $1
`

func (q *Queries) GetPostsForDid(ctx context.Context, did string) ([]Post, error) {
	rows, err := q.db.QueryContext(ctx, getPostsForDid, did)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Post
	for rows.Next() {
		var i Post
		if err := rows.Scan(
			&i.Did,
			&i.Rkey,
			&i.CreatedAt,
			&i.Text,
			&i.Reply,
			pq.Array(&i.Langs),
			&i.Likes,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const likePost = `-- name: LikePost :one
update posts
set likes = likes + 1
where did = $1
    and rkey = $2
returning did, rkey, created_at, text, reply, langs, likes
`

type LikePostParams struct {
	Did  string
	Rkey string
}

func (q *Queries) LikePost(ctx context.Context, arg LikePostParams) (Post, error) {
	row := q.db.QueryRowContext(ctx, likePost, arg.Did, arg.Rkey)
	var i Post
	err := row.Scan(
		&i.Did,
		&i.Rkey,
		&i.CreatedAt,
		&i.Text,
		&i.Reply,
		pq.Array(&i.Langs),
		&i.Likes,
	)
	return i, err
}
