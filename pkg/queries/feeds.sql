-- name: UpsertFeed :exec
insert into feeds (name, classifier)
values ($1, $2) on conflict (name) do
update
set classifier = excluded.classifier;
-- name: GetFeeds :many
select *
from feeds;