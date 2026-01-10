-- +goose Up
-- 1. Create the Event Container
CREATE TABLE events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL, -- e.g. "2026 Small Business Fair"
    event_date DATE NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- 2. Create a "Default" event to migrate existing tasks into
INSERT INTO events (name, event_date) 
VALUES ('General Tasks', NOW() + INTERVAL '1 year');

-- 3. Upgrade Tasks Table
ALTER TABLE tasks 
    ADD COLUMN event_id UUID REFERENCES events(id) ON DELETE CASCADE,
    ADD COLUMN category TEXT NOT NULL DEFAULT 'general', -- logistics, vendors, etc.
    ADD COLUMN completed_at TIMESTAMP, -- For "Soft Completion"
    ADD COLUMN is_archived BOOLEAN NOT NULL DEFAULT FALSE; -- For "Soft Delete"

-- 4. Migrate old data (Link orphans to the default event)
UPDATE tasks 
SET event_id = (SELECT id FROM events LIMIT 1) 
WHERE event_id IS NULL;

-- 5. Enforce Non-Null constraint after migration
ALTER TABLE tasks ALTER COLUMN event_id SET NOT NULL;

-- +goose Down
ALTER TABLE tasks DROP COLUMN is_archived;
ALTER TABLE tasks DROP COLUMN completed_at;
ALTER TABLE tasks DROP COLUMN category;
ALTER TABLE tasks DROP COLUMN event_id;
DROP TABLE events;