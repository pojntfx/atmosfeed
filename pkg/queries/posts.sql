-- name: CreatePost :exec
insert into posts (
        did,
        rkey,
        created_at,
        text,
        reply,
        langs,
        likes
    )
values ($1, $2, $3, $4, $5, $6, 0);
-- name: LikePost :one
update posts
set likes = likes + 1
where did = $1
    and rkey = $2
returning *;