// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.21.0
// source: feeds.sql

package models

import (
	"context"
	"time"
)

const deleteFeed = `-- name: DeleteFeed :exec
delete from feeds
where did = $1
    and rkey = $2
`

type DeleteFeedParams struct {
	Did  string
	Rkey string
}

func (q *Queries) DeleteFeed(ctx context.Context, arg DeleteFeedParams) error {
	_, err := q.db.ExecContext(ctx, deleteFeed, arg.Did, arg.Rkey)
	return err
}

const getFeedPosts = `-- name: GetFeedPosts :many
select p.did,
    p.rkey
from posts p
    join feed_posts fp on p.did = fp.post_did
    and p.rkey = fp.post_rkey
where fp.feed_did = $1
    and fp.feed_rkey = $2
    and p.created_at > $3
order by fp.weight desc
limit $4
`

type GetFeedPostsParams struct {
	FeedDid   string
	FeedRkey  string
	CreatedAt time.Time
	Limit     int32
}

type GetFeedPostsRow struct {
	Did  string
	Rkey string
}

func (q *Queries) GetFeedPosts(ctx context.Context, arg GetFeedPostsParams) ([]GetFeedPostsRow, error) {
	rows, err := q.db.QueryContext(ctx, getFeedPosts,
		arg.FeedDid,
		arg.FeedRkey,
		arg.CreatedAt,
		arg.Limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []GetFeedPostsRow
	for rows.Next() {
		var i GetFeedPostsRow
		if err := rows.Scan(&i.Did, &i.Rkey); err != nil {
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

const getFeedPostsCursor = `-- name: GetFeedPostsCursor :many
with referenceposttime as (
    select created_at
    from posts
    where posts.did = $5
        and posts.rkey = $6
)
select p.did,
    p.rkey
from posts p
    join feed_posts fp on p.did = fp.post_did
    and p.rkey = fp.post_rkey
where fp.feed_did = $1
    and fp.feed_rkey = $2
    and p.created_at > $3
    and p.created_at < (
        select created_at
        from referenceposttime
    )
order by fp.weight desc
limit $4
`

type GetFeedPostsCursorParams struct {
	FeedDid   string
	FeedRkey  string
	CreatedAt time.Time
	Limit     int32
	Did       string
	Rkey      string
}

type GetFeedPostsCursorRow struct {
	Did  string
	Rkey string
}

func (q *Queries) GetFeedPostsCursor(ctx context.Context, arg GetFeedPostsCursorParams) ([]GetFeedPostsCursorRow, error) {
	rows, err := q.db.QueryContext(ctx, getFeedPostsCursor,
		arg.FeedDid,
		arg.FeedRkey,
		arg.CreatedAt,
		arg.Limit,
		arg.Did,
		arg.Rkey,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []GetFeedPostsCursorRow
	for rows.Next() {
		var i GetFeedPostsCursorRow
		if err := rows.Scan(&i.Did, &i.Rkey); err != nil {
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

const getFeeds = `-- name: GetFeeds :many
select did, rkey
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
		if err := rows.Scan(&i.Did, &i.Rkey); err != nil {
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

const getFeedsForDid = `-- name: GetFeedsForDid :many
select did, rkey
from feeds
where did = $1
`

func (q *Queries) GetFeedsForDid(ctx context.Context, did string) ([]Feed, error) {
	rows, err := q.db.QueryContext(ctx, getFeedsForDid, did)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Feed
	for rows.Next() {
		var i Feed
		if err := rows.Scan(&i.Did, &i.Rkey); err != nil {
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
insert into feeds (did, rkey)
values ($1, $2) on conflict (did, rkey) do nothing
`

type UpsertFeedParams struct {
	Did  string
	Rkey string
}

func (q *Queries) UpsertFeed(ctx context.Context, arg UpsertFeedParams) error {
	_, err := q.db.ExecContext(ctx, upsertFeed, arg.Did, arg.Rkey)
	return err
}

const upsertFeedPost = `-- name: UpsertFeedPost :exec
insert into feed_posts (
        feed_did,
        feed_rkey,
        post_did,
        post_rkey,
        weight
    )
values ($1, $2, $3, $4, $5) on conflict (feed_did, feed_rkey, post_did, post_rkey) do
update
set weight = excluded.weight
`

type UpsertFeedPostParams struct {
	FeedDid  string
	FeedRkey string
	PostDid  string
	PostRkey string
	Weight   int32
}

func (q *Queries) UpsertFeedPost(ctx context.Context, arg UpsertFeedPostParams) error {
	_, err := q.db.ExecContext(ctx, upsertFeedPost,
		arg.FeedDid,
		arg.FeedRkey,
		arg.PostDid,
		arg.PostRkey,
		arg.Weight,
	)
	return err
}
