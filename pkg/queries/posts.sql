-- name: CreatePost :one
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
returning *;
-- name: LikePost :one
update posts
set likes = likes + 1
where did = $1
    and rkey = $2
returning *;
-- name: DeletePost :exec
delete from posts
where did = $1
    and rkey = $2;
-- name: DeleteAllPosts :exec
delete from posts;
-- name: GetPostsForDid :many
select *
from posts
where did = $1;
-- name: DeletePostsForDid :exec
delete from posts
where did = $1;