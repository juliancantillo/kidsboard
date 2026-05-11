-- name: GetActivityType :one
SELECT * FROM activity_types WHERE id = ?;

-- name: GetActivityTypeBySlug :one
SELECT * FROM activity_types WHERE slug = ? AND archived_at IS NULL;

-- name: ListActivityTypes :many
SELECT * FROM activity_types
WHERE archived_at IS NULL
ORDER BY category_id, name;

-- name: ListActivityTypesByCategory :many
SELECT * FROM activity_types
WHERE category_id = ? AND archived_at IS NULL
ORDER BY name;

-- name: UpsertActivityTypeBySlug :one
INSERT INTO activity_types (category_id, slug, name, description, xp_per_unit, points_per_unit)
VALUES (?, ?, ?, ?, ?, ?)
ON CONFLICT(slug) WHERE archived_at IS NULL
DO UPDATE SET
    category_id = excluded.category_id,
    name = excluded.name,
    description = excluded.description,
    xp_per_unit = excluded.xp_per_unit,
    points_per_unit = excluded.points_per_unit
RETURNING *;

-- name: CreateActivityType :one
INSERT INTO activity_types (category_id, slug, name, description, xp_per_unit, points_per_unit)
VALUES (?, ?, ?, ?, ?, ?)
RETURNING *;

-- name: UpdateActivityType :exec
UPDATE activity_types
SET category_id = ?, name = ?, description = ?, xp_per_unit = ?, points_per_unit = ?
WHERE id = ?;

-- name: ArchiveActivityType :exec
UPDATE activity_types SET archived_at = datetime('now') WHERE id = ? AND archived_at IS NULL;

-- name: UnarchiveActivityType :exec
UPDATE activity_types SET archived_at = NULL WHERE id = ?;
