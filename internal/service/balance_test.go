package service_test

import (
	"context"
	"testing"

	"cantillo.dev/kidsboard/internal/service"
	"cantillo.dev/kidsboard/internal/service/testutil"
	"cantillo.dev/kidsboard/internal/storage/sqldb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBalance_Earned_EmptyKid(t *testing.T) {
	ctx := context.Background()
	db := testutil.NewDB(t)
	q := sqldb.New(db)

	kid, err := q.CreateKid(ctx, sqldb.CreateKidParams{Name: "Mia", AvatarSlug: "mage", Color: "#FF00AA", DisplayOrder: 0})
	require.NoError(t, err)

	balance := service.NewBalanceService()
	earned, err := balance.Earned(ctx, db, kid.ID)
	require.NoError(t, err)
	assert.Equal(t, int64(0), earned)
}

func TestBalance_Spent_EmptyKidIsZero(t *testing.T) {
	ctx := context.Background()
	db := testutil.NewDB(t)
	q := sqldb.New(db)
	kid := testutil.NewKid(t, ctx, q, "Mia")

	balance := service.NewBalanceService()
	spent, err := balance.Spent(ctx, db, kid.ID)
	require.NoError(t, err)
	assert.Equal(t, int64(0), spent)
}

func TestBalance_PointsEarnedInCategory_OnlyCountsActivitiesInThatCategory(t *testing.T) {
	ctx := context.Background()
	db := testutil.NewDB(t)
	q := sqldb.New(db)
	kid := testutil.NewKid(t, ctx, q, "Mia")
	chores := testutil.NewCategory(t, ctx, q, "chores", "Chores")
	school := testutil.NewCategory(t, ctx, q, "school", "School")
	dishes := testutil.NewActivityType(t, ctx, q, chores.ID, "dishes", "Wash Dishes", 10, 5)
	homework := testutil.NewActivityType(t, ctx, q, school.ID, "homework", "Homework", 20, 10)

	// 3 dishes (15 pts), 2 homework (20 pts).
	for i := 0; i < 3; i++ {
		_, err := q.InsertActivity(ctx, sqldb.InsertActivityParams{
			KidID: kid.ID, ActivityTypeID: dishes.ID, Quantity: 1, XpAwarded: 10, PointsAwarded: 5,
		})
		require.NoError(t, err)
	}
	for i := 0; i < 2; i++ {
		_, err := q.InsertActivity(ctx, sqldb.InsertActivityParams{
			KidID: kid.ID, ActivityTypeID: homework.ID, Quantity: 1, XpAwarded: 20, PointsAwarded: 10,
		})
		require.NoError(t, err)
	}
	// Achievement bonus and adjustment — must NOT leak into category-scoped points.
	reason := "bonus"
	_, err := q.InsertPointAdjustment(ctx, sqldb.InsertPointAdjustmentParams{
		KidID: kid.ID, PointsDelta: 100, Reason: &reason,
	})
	require.NoError(t, err)

	balance := service.NewBalanceService()
	choresPts, err := balance.PointsEarnedInCategory(ctx, db, kid.ID, chores.ID)
	require.NoError(t, err)
	assert.Equal(t, int64(15), choresPts)

	schoolPts, err := balance.PointsEarnedInCategory(ctx, db, kid.ID, school.ID)
	require.NoError(t, err)
	assert.Equal(t, int64(20), schoolPts)
}

func TestBalance_Spent_CancelledRedemptionsReturnPoints(t *testing.T) {
	ctx := context.Background()
	db := testutil.NewDB(t)
	q := sqldb.New(db)
	kid := testutil.NewKid(t, ctx, q, "Mia")
	reward, err := q.CreateReward(ctx, sqldb.CreateRewardParams{
		Slug: "screen-time", Name: "30 min screen time", CostPoints: 50, Active: 1,
	})
	require.NoError(t, err)

	// Approve, then cancel — the points should no longer be Spent.
	red, err := q.InsertRedemption(ctx, sqldb.InsertRedemptionParams{
		KidID: kid.ID, RewardID: reward.ID, PointsSpent: 50,
	})
	require.NoError(t, err)
	require.NoError(t, q.TransitionRedemptionStatus(ctx, sqldb.TransitionRedemptionStatusParams{
		Status: "approved", ID: red.ID, Status_2: "pending",
	}))
	require.NoError(t, q.TransitionRedemptionStatus(ctx, sqldb.TransitionRedemptionStatusParams{
		Status: "cancelled", ID: red.ID, Status_2: "approved",
	}))

	balance := service.NewBalanceService()
	spent, err := balance.Spent(ctx, db, kid.ID)
	require.NoError(t, err)
	assert.Equal(t, int64(0), spent)
}

func TestBalance_Reserved_SumsPendingRedemptions(t *testing.T) {
	ctx := context.Background()
	db := testutil.NewDB(t)
	q := sqldb.New(db)
	kid := testutil.NewKid(t, ctx, q, "Mia")
	reward, err := q.CreateReward(ctx, sqldb.CreateRewardParams{
		Slug: "screen-time", Name: "30 min screen time", CostPoints: 50, Active: 1,
	})
	require.NoError(t, err)

	// Two pending: reserved totals 80.
	_, err = q.InsertRedemption(ctx, sqldb.InsertRedemptionParams{
		KidID: kid.ID, RewardID: reward.ID, PointsSpent: 30,
	})
	require.NoError(t, err)
	_, err = q.InsertRedemption(ctx, sqldb.InsertRedemptionParams{
		KidID: kid.ID, RewardID: reward.ID, PointsSpent: 50,
	})
	require.NoError(t, err)
	// One approved — does NOT count toward Reserved.
	approved, err := q.InsertRedemption(ctx, sqldb.InsertRedemptionParams{
		KidID: kid.ID, RewardID: reward.ID, PointsSpent: 100,
	})
	require.NoError(t, err)
	require.NoError(t, q.TransitionRedemptionStatus(ctx, sqldb.TransitionRedemptionStatusParams{
		Status: "approved", ID: approved.ID, Status_2: "pending",
	}))

	balance := service.NewBalanceService()
	reserved, err := balance.Reserved(ctx, db, kid.ID)
	require.NoError(t, err)
	assert.Equal(t, int64(80), reserved)
}

func TestBalance_Balance_EqualsEarnedMinusSpent(t *testing.T) {
	ctx := context.Background()
	db := testutil.NewDB(t)
	q := sqldb.New(db)
	kid := testutil.NewKid(t, ctx, q, "Mia")
	reason := "bonus"
	_, err := q.InsertPointAdjustment(ctx, sqldb.InsertPointAdjustmentParams{
		KidID: kid.ID, PointsDelta: 100, Reason: &reason,
	})
	require.NoError(t, err)
	penalty := "penalty"
	_, err = q.InsertPointAdjustment(ctx, sqldb.InsertPointAdjustmentParams{
		KidID: kid.ID, PointsDelta: -25, Reason: &penalty,
	})
	require.NoError(t, err)

	balance := service.NewBalanceService()
	val, err := balance.Balance(ctx, db, kid.ID)
	require.NoError(t, err)
	assert.Equal(t, int64(75), val) // 100 earned - 25 spent
}

func TestBalance_AvailableBalance_SubtractsReserved(t *testing.T) {
	ctx := context.Background()
	db := testutil.NewDB(t)
	q := sqldb.New(db)
	kid := testutil.NewKid(t, ctx, q, "Mia")
	reward, err := q.CreateReward(ctx, sqldb.CreateRewardParams{
		Slug: "screen-time", Name: "30 min screen time", CostPoints: 50, Active: 1,
	})
	require.NoError(t, err)

	bonus := "bonus"
	_, err = q.InsertPointAdjustment(ctx, sqldb.InsertPointAdjustmentParams{
		KidID: kid.ID, PointsDelta: 100, Reason: &bonus,
	})
	require.NoError(t, err)
	_, err = q.InsertRedemption(ctx, sqldb.InsertRedemptionParams{
		KidID: kid.ID, RewardID: reward.ID, PointsSpent: 30,
	})
	require.NoError(t, err)

	balance := service.NewBalanceService()
	avail, err := balance.AvailableBalance(ctx, db, kid.ID)
	require.NoError(t, err)
	assert.Equal(t, int64(70), avail) // 100 balance - 30 reserved
}

func TestBalance_Spent_CountsApprovedRedemptionsAndNegativeAdjustments(t *testing.T) {
	ctx := context.Background()
	db := testutil.NewDB(t)
	q := sqldb.New(db)
	kid := testutil.NewKid(t, ctx, q, "Mia")
	reward, err := q.CreateReward(ctx, sqldb.CreateRewardParams{
		Slug: "screen-time", Name: "30 min screen time", CostPoints: 50, Active: 1,
	})
	require.NoError(t, err)

	// Approved redemption — counts.
	approved, err := q.InsertRedemption(ctx, sqldb.InsertRedemptionParams{
		KidID: kid.ID, RewardID: reward.ID, PointsSpent: 50,
	})
	require.NoError(t, err)
	require.NoError(t, q.TransitionRedemptionStatus(ctx, sqldb.TransitionRedemptionStatusParams{
		Status: "approved", ID: approved.ID, Status_2: "pending",
	}))

	// Pending redemption — does NOT count toward Spent.
	_, err = q.InsertRedemption(ctx, sqldb.InsertRedemptionParams{
		KidID: kid.ID, RewardID: reward.ID, PointsSpent: 50,
	})
	require.NoError(t, err)

	// Negative adjustment — counts (as positive 7).
	reason := "broken vase"
	_, err = q.InsertPointAdjustment(ctx, sqldb.InsertPointAdjustmentParams{
		KidID: kid.ID, PointsDelta: -7, Reason: &reason,
	})
	require.NoError(t, err)

	balance := service.NewBalanceService()
	spent, err := balance.Spent(ctx, db, kid.ID)
	require.NoError(t, err)
	assert.Equal(t, int64(57), spent)
}

func TestBalance_Earned_IncludesPositiveAdjustmentsExcludesNegative(t *testing.T) {
	ctx := context.Background()
	db := testutil.NewDB(t)
	q := sqldb.New(db)

	kid := testutil.NewKid(t, ctx, q, "Mia")

	reasonBonus := "extra-credit"
	_, err := q.InsertPointAdjustment(ctx, sqldb.InsertPointAdjustmentParams{KidID: kid.ID, PointsDelta: 30, Reason: &reasonBonus})
	require.NoError(t, err)
	reasonPenalty := "broke vase"
	_, err = q.InsertPointAdjustment(ctx, sqldb.InsertPointAdjustmentParams{KidID: kid.ID, PointsDelta: -7, Reason: &reasonPenalty})
	require.NoError(t, err)

	balance := service.NewBalanceService()
	earned, err := balance.Earned(ctx, db, kid.ID)
	require.NoError(t, err)
	assert.Equal(t, int64(30), earned, "Earned counts the +30 bonus but not the -7 penalty")
}

func TestBalance_Earned_ExcludesVoidedActivities(t *testing.T) {
	ctx := context.Background()
	db := testutil.NewDB(t)
	q := sqldb.New(db)

	kid := testutil.NewKid(t, ctx, q, "Mia")
	cat := testutil.NewCategory(t, ctx, q, "chores", "Chores")
	at := testutil.NewActivityType(t, ctx, q, cat.ID, "wash-dishes", "Wash Dishes", 10, 5)

	a1, err := q.InsertActivity(ctx, sqldb.InsertActivityParams{
		KidID: kid.ID, ActivityTypeID: at.ID, Quantity: 1, XpAwarded: 10, PointsAwarded: 5,
	})
	require.NoError(t, err)
	_, err = q.InsertActivity(ctx, sqldb.InsertActivityParams{
		KidID: kid.ID, ActivityTypeID: at.ID, Quantity: 1, XpAwarded: 10, PointsAwarded: 5,
	})
	require.NoError(t, err)

	reason := "misclick"
	require.NoError(t, q.VoidActivity(ctx, sqldb.VoidActivityParams{ID: a1.ID, VoidReason: &reason}))

	balance := service.NewBalanceService()
	earned, err := balance.Earned(ctx, db, kid.ID)
	require.NoError(t, err)
	assert.Equal(t, int64(5), earned)
}

func TestBalance_Earned_IncludesActivityPoints(t *testing.T) {
	ctx := context.Background()
	db := testutil.NewDB(t)
	q := sqldb.New(db)

	kid := testutil.NewKid(t, ctx, q, "Mia")
	cat := testutil.NewCategory(t, ctx, q, "chores", "Chores")
	at := testutil.NewActivityType(t, ctx, q, cat.ID, "wash-dishes", "Wash Dishes", 10, 5)

	_, err := q.InsertActivity(ctx, sqldb.InsertActivityParams{
		KidID: kid.ID, ActivityTypeID: at.ID, Quantity: 1, XpAwarded: 10, PointsAwarded: 5,
	})
	require.NoError(t, err)
	_, err = q.InsertActivity(ctx, sqldb.InsertActivityParams{
		KidID: kid.ID, ActivityTypeID: at.ID, Quantity: 2, XpAwarded: 20, PointsAwarded: 10,
	})
	require.NoError(t, err)

	balance := service.NewBalanceService()
	earned, err := balance.Earned(ctx, db, kid.ID)
	require.NoError(t, err)
	assert.Equal(t, int64(15), earned)
}
