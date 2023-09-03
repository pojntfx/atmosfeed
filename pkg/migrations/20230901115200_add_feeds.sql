-- +goose Up
create table feeds (
    did text not null,
    rkey text not null,
    classifier bytea not null,
    primary key (did, rkey)
);
create table feed_posts (
    feed_did text not null,
    feed_rkey text not null,
    post_did text not null,
    post_rkey text not null,
    foreign key (feed_did, feed_rkey) references feeds(did, rkey) ON DELETE CASCADE,
    foreign key (post_did, post_rkey) references posts(did, rkey) ON DELETE CASCADE
);
-- +goose Down
drop table feed_posts;
drop table feeds;