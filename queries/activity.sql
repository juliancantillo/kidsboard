-- name: GetActivity :one
SELECT * FROM activities WHERE id = ?;

-- name: InsertActivity :one
INSERT INTO activities (
    kid_id, activity_type_id, quantity, xp_awarded, points_awarded, note, occurred_at
) VALUES (?, ?, ?, ?, ?, ?, ?)
RETURNING *;

-- name: VoidActivity :exec
UPDATE activities
SET voided_at = datetime('now'), void_reason = ?
WHERE id = ? AND voided_at IS NULL;

-- name: ListRecentActivitiesForKid :many
SELECT * FROM activities
WHERE kid_id = ? AND voided_at IS NULL
ORDER BY occurred_at DESC, id DESC
LIMIT ?;

-- name: ListActivitiesForKidSince :many
SELECT * FROM activities
WHERE kid_id = ?
  AND voided_at IS NULL
  AND occurred_at >= ?
ORDER BY occurred_at DESC, id DESC;

-- Engine support: count activities for a kid, optionally filtered by category.
-- Caller passes a non-NULL category_id to scope; NULL aggregates across all categories.
-- name: CountActivitiesForKid :one
SELECT COUNT(*) AS total
FROM activities a
JOIN activity_types t ON t.id = a.activity_type_id
WHERE a.kid_id = ?
  AND a.voided_at IS NULL
  AND (sqlc.narg('category_id') IS NULL OR t.category_id = sqlc.narg('category_id'));

-- Engine support: sum xp_awarded, optionally scoped to a category.
-- name: SumActivityXPForKid :one
SELECT CAST(COALESCE(SUM(a.xp_awarded), 0) AS INTEGER) AS total
FROM activities a
JOIN activity_types t ON t.id = a.activity_type_id
WHERE a.kid_id = ?
  AND a.voided_at IS NULL
  AND (sqlc.narg('category_id') IS NULL OR t.category_id = sqlc.narg('category_id'));

-- Engine + Balance support: sum points awarded from activities for a kid,
-- optionally scoped to a category. Category-scoped points rules use this directly.
-- name: SumActivityPointsForKid :one
SELECT CAST(COALESCE(SUM(a.points_awarded), 0) AS INTEGER) AS total
FROM activities a
JOIN activity_types t ON t.id = a.activity_type_id
WHERE a.kid_id = ?
  AND a.voided_at IS NULL
  AND (sqlc.narg('category_id') IS NULL OR t.category_id = sqlc.narg('category_id'));
