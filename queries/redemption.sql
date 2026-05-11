-- name: GetRedemption :one
SELECT * FROM redemptions WHERE id = ?;

-- name: InsertRedemption :one
INSERT INTO redemptions (kid_id, reward_id, points_spent, status)
VALUES (?, ?, ?, 'pending')
RETURNING *;

-- name: TransitionRedemptionStatus :exec
UPDATE redemptions
SET status = ?, decided_at = datetime('now')
WHERE id = ? AND status = ?;

-- name: ListPendingRedemptions :many
SELECT r.*, k.name AS kid_name, rw.name AS reward_name
FROM redemptions r
JOIN kids k ON k.id = r.kid_id
JOIN rewards rw ON rw.id = r.reward_id
WHERE r.status = 'pending'
ORDER BY r.requested_at;

-- name: ListPendingRedemptionsForKid :many
SELECT * FROM redemptions
WHERE kid_id = ? AND status = 'pending'
ORDER BY requested_at;

-- name: ListRedemptionsForKid :many
SELECT * FROM redemptions
WHERE kid_id = ?
ORDER BY requested_at DESC
LIMIT ?;

-- Balance support: sum approved redemption costs (the kid has "spent" these).
-- name: SumApprovedRedemptionPointsForKid :one
SELECT CAST(COALESCE(SUM(points_spent), 0) AS INTEGER) AS total
FROM redemptions
WHERE kid_id = ? AND status = 'approved';

-- Balance support: sum pending redemption costs (reserved against available balance).
-- name: SumPendingRedemptionPointsForKid :one
SELECT CAST(COALESCE(SUM(points_spent), 0) AS INTEGER) AS total
FROM redemptions
WHERE kid_id = ? AND status = 'pending';
