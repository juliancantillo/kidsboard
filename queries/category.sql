-- name: GetCategory :one
SELECT * FROM categories WHERE id = ?;

-- name: GetCategoryBySlug :one
SELECT * FROM categories WHERE slug = ? AND archived_at IS NULL;

-- name: ListCategories :many
SELECT * FROM categories
WHERE archived_at IS NULL
ORDER BY name;

-- name: CreateCategory :one
INSERT INTO categories (slug, name, description, icon, color)
VALUES (?, ?, ?, ?, ?)
RETURNING *;

-- name: UpdateCategory :exec
UPDATE categories
SET name = ?, description = ?, icon = ?, color = ?
WHERE id = ?;

-- name: UpsertCategoryBySlug :one
INSERT INTO categories (slug, name, description, icon, color)
VALUES (?, ?, ?, ?, ?)
ON CONFLICT(slug) WHERE archived_at IS NULL
DO UPDATE SET name = excluded.name, description = excluded.description,
              icon = excluded.icon, color = excluded.color
RETURNING *;

-- name: ArchiveCategory :exec
UPDATE categories SET archived_at = datetime('now') WHERE id = ? AND archived_at IS NULL;

-- name: UnarchiveCategory :exec
UPDATE categories SET archived_at = NULL WHERE id = ?;
