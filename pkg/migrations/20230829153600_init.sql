-- +goose Up
create table posts (
    did text not null,
    rkey text not null,
    created_at timestamp not null,
    text text not null,
    reply boolean not null,
    langs text [],
    likes int not null,
    primary key (did, rkey)
);
-- +goose Down
drop table posts;