-- +goose Up
-- +goose StatementBegin
-- Function for delete
create or replace function notify_feed_delete() returns trigger as $$ begin perform pg_notify('feed_deleted', old.did || '/' || old.rkey);
return old;
end;
$$ language plpgsql;
-- Trigger for delete
create trigger trigger_feed_delete
after delete on feeds for each row execute function notify_feed_delete();
-- +goose StatementEnd
-- +goose Down
drop trigger if exists trigger_feed_delete on feeds;
drop function if exists notify_feed_delete();