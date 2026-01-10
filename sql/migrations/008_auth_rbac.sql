-- +goose Up
-- 1. Upgrade People (Identity)
ALTER TABLE people ADD COLUMN email TEXT;
ALTER TABLE people ADD COLUMN password_hash TEXT;

-- Make email unique, but allow nulls for legacy/placeholder users initially
CREATE UNIQUE INDEX idx_people_email ON people(email) WHERE email IS NOT NULL;

-- 2. Event Membership (RBAC)
CREATE TABLE event_members (
    event_id UUID NOT NULL REFERENCES events(id) ON DELETE CASCADE,
    person_id UUID NOT NULL REFERENCES people(id) ON DELETE CASCADE,
    role TEXT NOT NULL CHECK (role IN ('owner', 'editor', 'viewer')),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    PRIMARY KEY (event_id, person_id)
);

-- 3. Audit Trail Actor
ALTER TABLE task_events ADD COLUMN actor_id UUID REFERENCES people(id);

-- 4. Session Store (for SCS library)
CREATE TABLE sessions (
    token TEXT PRIMARY KEY,
    data BYTEA NOT NULL,
    expiry TIMESTAMPTZ NOT NULL
);

CREATE INDEX sessions_expiry_idx ON sessions (expiry);

-- +goose Down
DROP TABLE sessions;
ALTER TABLE task_events DROP COLUMN actor_id;
DROP TABLE event_members;
ALTER TABLE people DROP COLUMN password_hash;
ALTER TABLE people DROP COLUMN email;