-- name: GetAchievement :one
SELECT * FROM achievements WHERE id = ?;

-- name: GetAchievementBySlug :one
SELECT * FROM achievements WHERE slug = ? AND archived_at IS NULL;

-- name: ListAchievements :many
SELECT * FROM achievements WHERE archived_at IS NULL ORDER BY name;

-- name: ListAllAchievements :many
SELECT * FROM achievements ORDER BY archived_at IS NOT NULL, name;

-- name: UpsertAchievementBySlug :one
INSERT INTO achievements (slug, name, description, title, combinator, bonus_points)
VALUES (?, ?, ?, ?, ?, ?)
ON CONFLICT(slug) WHERE archived_at IS NULL
DO UPDATE SET
    name = excluded.name,
    description = excluded.description,
    title = excluded.title,
    combinator = excluded.combinator,
    bonus_points = excluded.bonus_points
RETURNING *;

-- name: CreateAchievement :one
INSERT INTO achievements (slug, name, description, title, combinator, bonus_points)
VALUES (?, ?, ?, ?, ?, ?)
RETURNING *;

-- name: UpdateAchievement :exec
UPDATE achievements
SET name = ?, description = ?, title = ?, combinator = ?, bonus_points = ?
WHERE id = ?;

-- name: ArchiveAchievement :exec
UPDATE achievements SET archived_at = datetime('now') WHERE id = ? AND archived_at IS NULL;

-- name: UnarchiveAchievement :exec
UPDATE achievements SET archived_at = NULL WHERE id = ?;

-- name: InsertAchievementRule :one
INSERT INTO achievement_rules (achievement_id, category_id, metric, threshold)
VALUES (?, ?, ?, ?)
RETURNING *;

-- name: ListAchievementRules :many
SELECT * FROM achievement_rules WHERE achievement_id = ? ORDER BY id;

-- name: DeleteAchievementRules :exec
DELETE FROM achievement_rules WHERE achievement_id = ?;

-- Engine entry point: list every active achievement the kid has not yet earned,
-- joined to its rules so the engine can evaluate in a single round-trip.
-- Returns one row per rule; rule-less achievements are filtered out by the INNER JOIN.
-- name: ListUnearnedAchievementRulesForKid :many
SELECT
    a.id              AS achievement_id,
    a.slug            AS achievement_slug,
    a.name            AS achievement_name,
    a.description     AS achievement_description,
    a.title           AS achievement_title,
    a.combinator      AS achievement_combinator,
    a.bonus_points    AS achievement_bonus_points,
    a.archived_at     AS achievement_archived_at,
    a.created_at      AS achievement_created_at,
    r.id              AS rule_id,
    r.category_id     AS rule_category_id,
    r.metric          AS rule_metric,
    r.threshold       AS rule_threshold
FROM achievements a
INNER JOIN achievement_rules r ON r.achievement_id = a.id
WHERE a.archived_at IS NULL
  AND a.id NOT IN (
      SELECT achievement_id FROM kid_achievements WHERE kid_id = ?
  )
ORDER BY a.id, r.id;

-- name: MarkAchievementEarned :exec
INSERT INTO kid_achievements (kid_id, achievement_id, unseen)
VALUES (?, ?, 1)
ON CONFLICT(kid_id, achievement_id) DO NOTHING;

-- name: ListEarnedAchievementsForKid :many
SELECT
    ka.kid_id, ka.achievement_id, ka.earned_at, ka.unseen,
    a.slug, a.name, a.description, a.title, a.combinator,
    a.bonus_points, a.archived_at, a.created_at
FROM kid_achievements ka
JOIN achievements a ON a.id = ka.achievement_id
WHERE ka.kid_id = ?
ORDER BY ka.earned_at DESC;

-- name: MarkAchievementSeen :exec
UPDATE kid_achievements SET unseen = 0
WHERE kid_id = ? AND achievement_id = ?;

-- Balance support: sum of bonus_points across achievements the kid has earned.
-- name: SumEarnedAchievementBonusesForKid :one
SELECT CAST(COALESCE(SUM(a.bonus_points), 0) AS INTEGER) AS total
FROM kid_achievements ka
JOIN achievements a ON a.id = ka.achievement_id
WHERE ka.kid_id = ?;
