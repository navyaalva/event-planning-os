-- +goose Up
-- Ensure uuid generator exists
CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE task_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    task_id UUID NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
    event_type TEXT NOT NULL, -- 'CREATED', 'UPDATED', 'COMMENT'
    changes JSONB NOT NULL DEFAULT '{}'::jsonb, -- Diff payload
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_task_events_task_id ON task_events(task_id);

-- +goose Down
DROP TABLE IF EXISTS task_events;
