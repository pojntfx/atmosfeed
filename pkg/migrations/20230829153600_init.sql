-- +goose Up
create table posts (
    id serial primary key,
    created_at timestamp not null,
    did text not null,
    rkey text not null,
    text text not null,
    reply boolean not null,
    langs text [] not null,
    likes int not null
);
-- +goose Down
drop table posts;