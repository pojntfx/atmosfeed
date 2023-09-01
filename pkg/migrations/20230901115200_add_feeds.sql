-- +goose Up
create table feeds (
    name text not null primary key,
    classifier bytea not null
);
create table feed_posts (
    feed_name text not null,
    post_did text not null,
    post_rkey text not null,
    foreign key (feed_name) references feeds(name) ON DELETE CASCADE,
    foreign key (post_did, post_rkey) references posts(did, rkey) ON DELETE CASCADE
);
-- +goose Down
drop table feed_posts;
drop table feeds;