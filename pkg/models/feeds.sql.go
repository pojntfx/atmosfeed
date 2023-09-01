// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.20.0
// source: feeds.sql

package models

import (
	"context"
)

const createFeedPost = `-- name: CreateFeedPost :exec
insert into feed_posts (
        feed_name,
        post_did,
        post_rkey
    )
values ($1, $2, $3)
`

type CreateFeedPostParams struct {
	FeedName string
	PostDid  string
	PostRkey string
}

func (q *Queries) CreateFeedPost(ctx context.Context, arg CreateFeedPostParams) error {
	_, err := q.db.ExecContext(ctx, createFeedPost, arg.FeedName, arg.PostDid, arg.PostRkey)
	return err
}

const deleteFeed = `-- name: DeleteFeed :exec
delete from feeds
where name = $1
`

func (q *Queries) DeleteFeed(ctx context.Context, name string) error {
	_, err := q.db.ExecContext(ctx, deleteFeed, name)
	return err
}

const getFeedClassifier = `-- name: GetFeedClassifier :one
select classifier
from feeds
where name = $1
`

func (q *Queries) GetFeedClassifier(ctx context.Context, name string) ([]byte, error) {
	row := q.db.QueryRowContext(ctx, getFeedClassifier, name)
	var classifier []byte
	err := row.Scan(&classifier)
	return classifier, err
}

const getFeeds = `-- name: GetFeeds :many
select name, classifier
from feeds
`

func (q *Queries) GetFeeds(ctx context.Context) ([]Feed, error) {
	rows, err := q.db.QueryContext(ctx, getFeeds)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Feed
	for rows.Next() {
		var i Feed
		if err := rows.Scan(&i.Name, &i.Classifier); err != nil {
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

const upsertFeed = `-- name: UpsertFeed :exec
insert into feeds (name, classifier)
values ($1, $2) on conflict (name) do
update
set classifier = excluded.classifier
`

type UpsertFeedParams struct {
	Name       string
	Classifier []byte
}

func (q *Queries) UpsertFeed(ctx context.Context, arg UpsertFeedParams) error {
	_, err := q.db.ExecContext(ctx, upsertFeed, arg.Name, arg.Classifier)
	return err
}
