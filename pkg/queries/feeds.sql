-- name: UpsertFeed :exec
insert into feeds (did, rkey, classifier)
values ($1, $2, $3) on conflict (did, rkey) do
update
set classifier = excluded.classifier;
-- name: GetFeeds :many
select *
from feeds;
-- name: GetFeedClassifier :one
select classifier
from feeds
where did = $1
    and rkey = $2;
-- name: DeleteFeed :exec
delete from feeds
where did = $1
    and rkey = $2;
-- name: UpsertFeedPost :exec
insert into feed_posts (
        feed_did,
        feed_rkey,
        post_did,
        post_rkey,
        weight
    )
values ($1, $2, $3, $4, $5) on conflict (feed_did, feed_rkey, post_did, post_rkey) do
update
set weight = excluded.weight;
-- name: GetFeedPosts :many
select p.did,
    p.rkey
from posts p
    join feed_posts fp on p.did = fp.post_did
    and p.rkey = fp.post_rkey
where fp.feed_did = $1
    and fp.feed_rkey = $2
    and p.created_at > $3
order by fp.weight desc
limit $4;
-- name: GetFeedPostsCursor :many
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
limit $4;