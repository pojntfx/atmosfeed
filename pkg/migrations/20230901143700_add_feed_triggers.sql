-- +goose Up
-- +goose StatementBegin
-- Function for insert
create or replace function notify_feed_insert() returns trigger as $$ begin perform pg_notify('feed_inserted', new.did || '/' || new.rkey);
return new;
end;
$$ language plpgsql;
-- Function for update
create or replace function notify_feed_update() returns trigger as $$ begin perform pg_notify('feed_updated', old.did || '/' || new.rkey);
return new;
end;
$$ language plpgsql;
-- Function for delete
create or replace function notify_feed_delete() returns trigger as $$ begin perform pg_notify('feed_deleted', old.did || '/' || old.rkey);
return old;
end;
$$ language plpgsql;
-- Trigger for insert
create trigger trigger_feed_insert
after
insert on feeds for each row execute function notify_feed_insert();
-- Trigger for update
create trigger trigger_feed_update
after
update on feeds for each row execute function notify_feed_update();
-- Trigger for delete
create trigger trigger_feed_delete
after delete on feeds for each row execute function notify_feed_delete();
-- +goose StatementEnd
-- +goose Down
drop trigger if exists trigger_feed_insert on feeds;
drop function if exists notify_feed_insert();
drop trigger if exists trigger_feed_update on feeds;
drop function if exists notify_feed_update();
drop trigger if exists trigger_feed_delete on feeds;
drop function if exists notify_feed_delete();