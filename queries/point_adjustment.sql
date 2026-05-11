-- name: GetPointAdjustment :one
SELECT * FROM point_adjustments WHERE id = ?;

-- name: InsertPointAdjustment :one
INSERT INTO point_adjustments (kid_id, points_delta, reason)
VALUES (?, ?, ?)
RETURNING *;

-- name: VoidPointAdjustment :exec
UPDATE point_adjustments
SET voided_at = datetime('now'), void_reason = ?
WHERE id = ? AND voided_at IS NULL;

-- name: ListPointAdjustmentsForKid :many
SELECT * FROM point_adjustments
WHERE kid_id = ?
ORDER BY created_at DESC
LIMIT ?;

-- Balance support: signed sum of non-voided adjustments for a kid.
-- A positive total adds to Earned (when delta > 0) or to Spent (when delta < 0);
-- the BalanceCalculator splits this in Go.
-- name: SumPointAdjustmentsForKid :one
SELECT CAST(COALESCE(SUM(points_delta), 0) AS INTEGER) AS total
FROM point_adjustments
WHERE kid_id = ? AND voided_at IS NULL;

-- Balance support: separate positive (earned) and negative (spent) totals.
-- name: SumPositivePointAdjustmentsForKid :one
SELECT CAST(COALESCE(SUM(points_delta), 0) AS INTEGER) AS total
FROM point_adjustments
WHERE kid_id = ? AND voided_at IS NULL AND points_delta > 0;

-- name: SumNegativePointAdjustmentsForKid :one
SELECT CAST(COALESCE(-SUM(points_delta), 0) AS INTEGER) AS total
FROM point_adjustments
WHERE kid_id = ? AND voided_at IS NULL AND points_delta < 0;
