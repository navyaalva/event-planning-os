-- +goose Up
CREATE TABLE templates (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL,
    description TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TABLE template_tasks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    template_id UUID NOT NULL REFERENCES templates(id) ON DELETE CASCADE,
    title TEXT NOT NULL,
    description TEXT,
    category TEXT NOT NULL DEFAULT 'general',
    priority INT NOT NULL DEFAULT 3,
    relative_due_days INT DEFAULT 0 -- e.g., 7 days before event
);

-- Seed Data
INSERT INTO templates (id, name, description) 
VALUES ('11111111-1111-1111-1111-111111111111', 'Vendor Fair Standard', 'Default checklist for fairs');

INSERT INTO template_tasks (template_id, title, category, priority, relative_due_days) VALUES
('11111111-1111-1111-1111-111111111111', 'Secure Venue', 'logistics', 5, 90),
('11111111-1111-1111-1111-111111111111', 'Create Vendor Form', 'vendors', 4, 60),
('11111111-1111-1111-1111-111111111111', 'Send Invoices', 'finance', 5, 14);

-- +goose Down
DROP TABLE template_tasks;
DROP TABLE templates;