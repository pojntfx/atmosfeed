-- name: CreatePost :one
insert into posts (created_at, did, rkey, text, reply, langs, likes)
values ($1, $2, $3, $4, $5, $6, 0)
returning id;