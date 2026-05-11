-- name: GetReward :one
SELECT * FROM rewards WHERE id = ?;

-- name: GetRewardBySlug :one
SELECT * FROM rewards WHERE slug = ? AND archived_at IS NULL;

-- name: ListRewards :many
SELECT * FROM rewards WHERE archived_at IS NULL ORDER BY cost_points;

-- name: ListActiveRewards :many
SELECT * FROM rewards
WHERE archived_at IS NULL AND active = 1
ORDER BY cost_points;

-- name: CreateReward :one
INSERT INTO rewards (slug, name, description, cost_points, active)
VALUES (?, ?, ?, ?, ?)
RETURNING *;

-- name: UpdateReward :exec
UPDATE rewards
SET name = ?, description = ?, cost_points = ?, active = ?
WHERE id = ?;

-- name: ArchiveReward :exec
UPDATE rewards SET archived_at = datetime('now') WHERE id = ? AND archived_at IS NULL;

-- name: UnarchiveReward :exec
UPDATE rewards SET archived_at = NULL WHERE id = ?;
