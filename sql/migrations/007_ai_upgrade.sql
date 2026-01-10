-- +goose Up
ALTER TABLE tasks ADD COLUMN assignee_text TEXT;
ALTER TABLE tasks ADD COLUMN subtasks JSONB DEFAULT '[]';

-- +goose Down
ALTER TABLE tasks DROP COLUMN subtasks;
ALTER TABLE tasks DROP COLUMN assignee_text;