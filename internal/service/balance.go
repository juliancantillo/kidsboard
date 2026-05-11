package service

import (
	"context"

	"cantillo.dev/kidsboard/internal/storage"
	"cantillo.dev/kidsboard/internal/storage/sqldb"
)

// BalanceService derives a kid's points figures from the event tables.
// It performs no mutation; all results are computed live.
type BalanceService interface {
	Earned(ctx context.Context, db storage.DBTX, kidID int64) (int64, error)
	Spent(ctx context.Context, db storage.DBTX, kidID int64) (int64, error)
	Reserved(ctx context.Context, db storage.DBTX, kidID int64) (int64, error)
	Balance(ctx context.Context, db storage.DBTX, kidID int64) (int64, error)
	AvailableBalance(ctx context.Context, db storage.DBTX, kidID int64) (int64, error)
	PointsEarnedInCategory(ctx context.Context, db storage.DBTX, kidID, categoryID int64) (int64, error)
}

type balanceService struct{}

func NewBalanceService() BalanceService { return &balanceService{} }

// Earned is the sum of every points-positive source for a kid:
// activity points + earned achievement bonuses + positive (non-voided) adjustments.
func (s *balanceService) Earned(ctx context.Context, db storage.DBTX, kidID int64) (int64, error) {
	q := sqldb.New(db)
	activityPoints, err := q.SumActivityPointsForKid(ctx, sqldb.SumActivityPointsForKidParams{KidID: kidID, CategoryID: nil})
	if err != nil {
		return 0, err
	}
	bonuses, err := q.SumEarnedAchievementBonusesForKid(ctx, kidID)
	if err != nil {
		return 0, err
	}
	posAdj, err := q.SumPositivePointAdjustmentsForKid(ctx, kidID)
	if err != nil {
		return 0, err
	}
	return activityPoints + bonuses + posAdj, nil
}

// Spent is the sum of every points-negative source: approved redemptions
// plus negative (non-voided) adjustments expressed as positive numbers.
func (s *balanceService) Spent(ctx context.Context, db storage.DBTX, kidID int64) (int64, error) {
	q := sqldb.New(db)
	approved, err := q.SumApprovedRedemptionPointsForKid(ctx, kidID)
	if err != nil {
		return 0, err
	}
	negAdj, err := q.SumNegativePointAdjustmentsForKid(ctx, kidID)
	if err != nil {
		return 0, err
	}
	return approved + negAdj, nil
}

// Reserved is the sum of pending redemption costs — points the kid has
// asked to spend but the parent has not yet decided on.
func (s *balanceService) Reserved(ctx context.Context, db storage.DBTX, kidID int64) (int64, error) {
	return sqldb.New(db).SumPendingRedemptionPointsForKid(ctx, kidID)
}

// Balance is Earned minus Spent — what the kid "has" right now,
// not counting points encumbered by pending redemption requests.
func (s *balanceService) Balance(ctx context.Context, db storage.DBTX, kidID int64) (int64, error) {
	earned, err := s.Earned(ctx, db, kidID)
	if err != nil {
		return 0, err
	}
	spent, err := s.Spent(ctx, db, kidID)
	if err != nil {
		return 0, err
	}
	return earned - spent, nil
}

// AvailableBalance is Balance minus Reserved — what the kid can request
// to spend right now. The shop uses this to gate new redemption requests.
func (s *balanceService) AvailableBalance(ctx context.Context, db storage.DBTX, kidID int64) (int64, error) {
	balance, err := s.Balance(ctx, db, kidID)
	if err != nil {
		return 0, err
	}
	reserved, err := s.Reserved(ctx, db, kidID)
	if err != nil {
		return 0, err
	}
	return balance - reserved, nil
}

// PointsEarnedInCategory returns the sum of points awarded by activities
// in a single category. Category-scoped points rules use this exclusively —
// achievement bonuses and adjustments are NOT category-scoped and do not
// contribute. For a global points figure, use Earned/Balance instead.
func (s *balanceService) PointsEarnedInCategory(ctx context.Context, db storage.DBTX, kidID, categoryID int64) (int64, error) {
	return sqldb.New(db).SumActivityPointsForKid(ctx, sqldb.SumActivityPointsForKidParams{
		KidID:      kidID,
		CategoryID: categoryID,
	})
}
