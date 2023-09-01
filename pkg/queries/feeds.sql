-- name: UpsertFeed :exec
insert into feeds (name, classifier)
values ($1, $2) on conflict (name) do
update
set classifier = excluded.classifier;
-- name: GetFeeds :many
select *
from feeds;
-- name: GetFeedClassifier :one
select classifier
from feeds
where name = $1;
-- name: DeleteFeed :exec
delete from feeds
where name = $1;
-- name: CreateFeedPost :exec
insert into feed_posts (
        feed_name,
        post_did,
        post_rkey
    )
values ($1, $2, $3);