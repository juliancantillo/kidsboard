-- name: GetKid :one
SELECT * FROM kids WHERE id = ?;

-- name: ListKids :many
SELECT * FROM kids
WHERE (sqlc.narg('include_archived') IS NOT NULL AND sqlc.narg('include_archived') = 1)
   OR archived_at IS NULL
ORDER BY display_order, name;

-- name: CreateKid :one
INSERT INTO kids (name, avatar_slug, color, display_order)
VALUES (?, ?, ?, ?)
RETURNING *;

-- name: UpdateKid :exec
UPDATE kids
SET name = ?, avatar_slug = ?, color = ?, display_order = ?
WHERE id = ?;

-- name: ArchiveKid :exec
UPDATE kids SET archived_at = datetime('now') WHERE id = ? AND archived_at IS NULL;

-- name: UnarchiveKid :exec
UPDATE kids SET archived_at = NULL WHERE id = ?;
