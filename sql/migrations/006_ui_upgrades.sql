-- +goose Up
ALTER TABLE events ADD COLUMN location TEXT;
ALTER TABLE events ADD COLUMN summary TEXT;
ALTER TABLE tasks ADD COLUMN deleted_at TIMESTAMP;

-- +goose Down
ALTER TABLE tasks DROP COLUMN deleted_at;
ALTER TABLE events DROP COLUMN summary;
ALTER TABLE events DROP COLUMN location;