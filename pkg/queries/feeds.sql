-- name: UpsertFeedMetadata :exec
insert into feeds (did, rkey, pinned_did, pinned_rkey)
values ($1, $2, $3, $4) on conflict (did, rkey) do
update
set pinned_did = excluded.pinned_did,
    pinned_rkey = excluded.pinned_rkey;
-- name: UpsertFeedClassifier :exec
insert into feeds (did, rkey, pinned_did, pinned_rkey)
values ($1, $2, '', '') on conflict (did, rkey) do nothing;
-- name: GetFeeds :many
select *
from feeds;
-- name: GetFeedsForDid :many
select *
from feeds
where did = $1;
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
with pinned_post as (
    select pinned_did,
        pinned_rkey
    from feeds f
    where f.did = $1
        and f.rkey = $2
),
feed_posts as (
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
)
select did,
    rkey
from (
        select pinned_did as did,
            pinned_rkey as rkey
        from pinned_post p
        where p.pinned_did <> ''
            and p.pinned_rkey <> ''
        union all
        select did,
            rkey
        from feed_posts
    ) as results;
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
-- name: GetFeedPostsForDid :many
select *
from feed_posts
where post_did = $1;
-- name: DeleteFeedPostsForDid :exec
delete from feed_posts
where post_did = $1;