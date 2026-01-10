-- +goose Up
CREATE TABLE task_dependencies (
    task_id UUID NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
    dependency_id UUID NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    PRIMARY KEY (task_id, dependency_id),
    CONSTRAINT no_self_dependency CHECK (task_id != dependency_id)
);

CREATE INDEX idx_task_deps_dependency ON task_dependencies(dependency_id);

-- +goose Down
DROP TABLE task_dependencies;