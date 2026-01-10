-- name: ListEvents :many
SELECT 
    e.id, 
    e.name, 
    e.event_date,
    e.location,
    e.summary,
    COUNT(t.id) as total_tasks, 
    COUNT(t.id) FILTER (WHERE t.status = 'done' AND t.deleted_at IS NULL) as completed_tasks
FROM events e
LEFT JOIN tasks t ON e.id = t.event_id AND t.deleted_at IS NULL
GROUP BY e.id
ORDER BY e.event_date ASC;

-- name: GetEventTasks :many
SELECT 
    t.*, 
    p.name as owner_name 
FROM tasks t
LEFT JOIN people p ON t.owner_id = p.id
WHERE t.event_id = $1 
AND t.deleted_at IS NULL
AND ($2::boolean = TRUE OR t.status != 'done')
ORDER BY 
    CASE WHEN t.status = 'done' THEN 1 ELSE 0 END,
    t.due_date ASC NULLS LAST;

-- name: SoftDeleteTask :exec
UPDATE tasks 
SET deleted_at = NOW() 
WHERE id = $1;

-- name: CreateTask :one
INSERT INTO tasks (
    title, description, owner_id, priority, due_date, tags, event_id, category,
    assignee_text, subtasks
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10) 
RETURNING *;

-- name: GetTask :one
SELECT * FROM tasks WHERE id = $1 AND deleted_at IS NULL;

-- name: UpdateTask :one
UPDATE tasks
SET 
    title       = COALESCE(sqlc.narg(title), title),
    description = COALESCE(sqlc.narg(description), description),
    status      = COALESCE(sqlc.narg(status), status),
    priority    = COALESCE(sqlc.narg(priority), priority),
    due_date    = COALESCE(sqlc.narg(due_date), due_date),
    category    = COALESCE(sqlc.narg(category), category),
    owner_id    = COALESCE(sqlc.narg(owner_id), owner_id),
    assignee_text = COALESCE(sqlc.narg(assignee_text), assignee_text), -- Added
    subtasks    = COALESCE(sqlc.narg(subtasks), subtasks),             -- Added
    last_update_at = NOW()
WHERE id = sqlc.arg(id)
RETURNING *;

-- name: CreateTaskEvent :exec
INSERT INTO task_events (task_id, event_type, changes)
VALUES ($1, $2, $3);

-- name: GetTaskEvents :many
SELECT * FROM task_events WHERE task_id = $1 ORDER BY created_at DESC;

-- name: ListPeople :many
SELECT * FROM people ORDER BY name ASC;

-- name: GetPerson :one
SELECT * FROM people WHERE id = $1;

-- name: ListTemplates :many
SELECT * FROM templates ORDER BY name ASC;

-- name: GetTemplateTasks :many
SELECT * FROM template_tasks WHERE template_id = $1;

-- name: CreateEvent :one
INSERT INTO events (name, event_date) VALUES ($1, $2) RETURNING *;

-- name: GetEvent :one
SELECT * FROM events WHERE id = $1;

-- name: UpdateEvent :one
UPDATE events
SET 
    name = COALESCE(sqlc.narg(name), name),
    event_date = COALESCE(sqlc.narg(event_date), event_date),
    location = COALESCE(sqlc.narg(location), location),
    summary = COALESCE(sqlc.narg(summary), summary)
WHERE id = sqlc.arg(id)
RETURNING *;

-- name: BatchSoftDeleteTasks :exec
UPDATE tasks 
SET deleted_at = NOW() 
WHERE id = ANY($1::uuid[]);

-- name: GetTasksForFollowUp :many
SELECT * FROM tasks 
WHERE status != 'done' 
AND deleted_at IS NULL
AND due_date IS NOT NULL 
AND assignee_text IS NOT NULL 
AND assignee_text != '';

-- name: GetGlobalActiveTasks :many
SELECT 
    t.*, 
    p.name as owner_name,
    e.name as event_name
FROM tasks t
LEFT JOIN people p ON t.owner_id = p.id
JOIN events e ON t.event_id = e.id
WHERE t.status != 'done' 
AND t.deleted_at IS NULL
ORDER BY t.priority DESC, t.due_date ASC;

-- name: GetPersonByEmail :one
SELECT * FROM people WHERE email = $1;

-- name: CreatePerson :one
INSERT INTO people (name, email, password_hash, role)
VALUES ($1, $2, $3, 'user')
RETURNING *;

-- name: AddEventMember :exec
INSERT INTO event_members (event_id, person_id, role)
VALUES ($1, $2, $3);

-- name: ListUserEvents :many
SELECT 
    e.id, e.name, e.event_date, e.location, e.summary,
    em.role as user_role,
    COUNT(t.id) as total_tasks, 
    COUNT(t.id) FILTER (WHERE t.status = 'done' AND t.deleted_at IS NULL) as completed_tasks
FROM events e
JOIN event_members em ON e.id = em.event_id
LEFT JOIN tasks t ON e.id = t.event_id AND t.deleted_at IS NULL
WHERE em.person_id = $1
GROUP BY e.id, em.role
ORDER BY e.event_date ASC;

-- name: GetEventMembership :one
SELECT role FROM event_members 
WHERE event_id = $1 AND person_id = $2;