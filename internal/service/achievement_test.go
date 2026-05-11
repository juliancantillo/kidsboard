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

func TestAchievement_Reevaluate_ThresholdNotMet(t *testing.T) {
	ctx := context.Background()
	db := testutil.NewDB(t)
	q := sqldb.New(db)

	kid := testutil.NewKid(t, ctx, q, "Mia")
	chores := testutil.NewCategory(t, ctx, q, "chores", "Chores")
	dishes := testutil.NewActivityType(t, ctx, q, chores.ID, "wash-dishes", "Wash Dishes", 10, 5)
	testutil.NewAchievement(t, ctx, q, testutil.AchievementSpec{
		Slug: "dish-champ", Name: "Dish Champion", Combinator: "ALL",
		Rules: []testutil.RuleSpec{{CategoryID: &chores.ID, Metric: "count", Threshold: 3}},
	})

	testutil.LogActivity(t, ctx, q, kid.ID, dishes.ID, 1) // only 1 of 3

	engine := service.NewAchievementService(service.NewBalanceService())
	earned, err := engine.Reevaluate(ctx, db, kid.ID)
	require.NoError(t, err)
	assert.Empty(t, earned)
}

func TestAchievement_Reevaluate_AllCombinatorRequiresAllRules(t *testing.T) {
	ctx := context.Background()
	db := testutil.NewDB(t)
	q := sqldb.New(db)

	kid := testutil.NewKid(t, ctx, q, "Mia")
	chores := testutil.NewCategory(t, ctx, q, "chores", "Chores")
	school := testutil.NewCategory(t, ctx, q, "school", "School")
	dishes := testutil.NewActivityType(t, ctx, q, chores.ID, "wash-dishes", "Wash Dishes", 10, 5)
	homework := testutil.NewActivityType(t, ctx, q, school.ID, "homework", "Homework", 20, 10)

	testutil.NewAchievement(t, ctx, q, testutil.AchievementSpec{
		Slug: "renaissance", Name: "Renaissance Kid", Combinator: "ALL",
		Rules: []testutil.RuleSpec{
			{CategoryID: &chores.ID, Metric: "count", Threshold: 3},
			{CategoryID: &school.ID, Metric: "count", Threshold: 3},
		},
	})

	for i := 0; i < 3; i++ {
		testutil.LogActivity(t, ctx, q, kid.ID, dishes.ID, 1)
	}
	testutil.LogActivity(t, ctx, q, kid.ID, homework.ID, 1) // only 1 homework

	engine := service.NewAchievementService(service.NewBalanceService())
	earned, err := engine.Reevaluate(ctx, db, kid.ID)
	require.NoError(t, err)
	assert.Empty(t, earned, "ALL needs both rules to pass; homework rule failed")
}

func TestAchievement_Reevaluate_AnyCombinatorOneRulePasses(t *testing.T) {
	ctx := context.Background()
	db := testutil.NewDB(t)
	q := sqldb.New(db)

	kid := testutil.NewKid(t, ctx, q, "Mia")
	chores := testutil.NewCategory(t, ctx, q, "chores", "Chores")
	school := testutil.NewCategory(t, ctx, q, "school", "School")
	dishes := testutil.NewActivityType(t, ctx, q, chores.ID, "wash-dishes", "Wash Dishes", 10, 5)
	testutil.NewActivityType(t, ctx, q, school.ID, "homework", "Homework", 20, 10)

	testutil.NewAchievement(t, ctx, q, testutil.AchievementSpec{
		Slug: "helper", Name: "Helper", Combinator: "ANY",
		Rules: []testutil.RuleSpec{
			{CategoryID: &chores.ID, Metric: "count", Threshold: 3},
			{CategoryID: &school.ID, Metric: "count", Threshold: 3},
		},
	})

	for i := 0; i < 3; i++ {
		testutil.LogActivity(t, ctx, q, kid.ID, dishes.ID, 1)
	}

	engine := service.NewAchievementService(service.NewBalanceService())
	earned, err := engine.Reevaluate(ctx, db, kid.ID)
	require.NoError(t, err)
	require.Len(t, earned, 1)
	assert.Equal(t, "Helper", earned[0].Name)
}

func TestAchievement_Reevaluate_IsIdempotent(t *testing.T) {
	ctx := context.Background()
	db := testutil.NewDB(t)
	q := sqldb.New(db)

	kid := testutil.NewKid(t, ctx, q, "Mia")
	chores := testutil.NewCategory(t, ctx, q, "chores", "Chores")
	dishes := testutil.NewActivityType(t, ctx, q, chores.ID, "wash-dishes", "Wash Dishes", 10, 5)
	testutil.NewAchievement(t, ctx, q, testutil.AchievementSpec{
		Slug: "dish-champ", Name: "Dish Champion", Combinator: "ALL",
		Rules: []testutil.RuleSpec{{CategoryID: &chores.ID, Metric: "count", Threshold: 3}},
	})
	for i := 0; i < 3; i++ {
		testutil.LogActivity(t, ctx, q, kid.ID, dishes.ID, 1)
	}

	engine := service.NewAchievementService(service.NewBalanceService())
	first, err := engine.Reevaluate(ctx, db, kid.ID)
	require.NoError(t, err)
	require.Len(t, first, 1)

	second, err := engine.Reevaluate(ctx, db, kid.ID)
	require.NoError(t, err)
	assert.Empty(t, second, "second call must not re-fire any earned achievement")
}

func TestAchievement_Reevaluate_VoidedActivityDoesNotCount(t *testing.T) {
	ctx := context.Background()
	db := testutil.NewDB(t)
	q := sqldb.New(db)

	kid := testutil.NewKid(t, ctx, q, "Mia")
	chores := testutil.NewCategory(t, ctx, q, "chores", "Chores")
	dishes := testutil.NewActivityType(t, ctx, q, chores.ID, "wash-dishes", "Wash Dishes", 10, 5)
	testutil.NewAchievement(t, ctx, q, testutil.AchievementSpec{
		Slug: "dish-champ", Name: "Dish Champion", Combinator: "ALL",
		Rules: []testutil.RuleSpec{{CategoryID: &chores.ID, Metric: "count", Threshold: 3}},
	})

	// Log 3 activities then void one — only 2 should count.
	var first sqldb.Activity
	for i := 0; i < 3; i++ {
		act := testutil.LogActivity(t, ctx, q, kid.ID, dishes.ID, 1)
		if i == 0 {
			first = act
		}
	}
	reason := "misclick"
	require.NoError(t, q.VoidActivity(ctx, sqldb.VoidActivityParams{ID: first.ID, VoidReason: &reason}))

	engine := service.NewAchievementService(service.NewBalanceService())
	earned, err := engine.Reevaluate(ctx, db, kid.ID)
	require.NoError(t, err)
	assert.Empty(t, earned)
}

func TestAchievement_Reevaluate_AchievementStaysEarnedAfterUnderlyingActivityVoided(t *testing.T) {
	ctx := context.Background()
	db := testutil.NewDB(t)
	q := sqldb.New(db)

	kid := testutil.NewKid(t, ctx, q, "Mia")
	chores := testutil.NewCategory(t, ctx, q, "chores", "Chores")
	dishes := testutil.NewActivityType(t, ctx, q, chores.ID, "wash-dishes", "Wash Dishes", 10, 5)
	testutil.NewAchievement(t, ctx, q, testutil.AchievementSpec{
		Slug: "dish-champ", Name: "Dish Champion", Combinator: "ALL",
		Rules: []testutil.RuleSpec{{CategoryID: &chores.ID, Metric: "count", Threshold: 3}},
	})

	var first sqldb.Activity
	for i := 0; i < 3; i++ {
		act := testutil.LogActivity(t, ctx, q, kid.ID, dishes.ID, 1)
		if i == 0 {
			first = act
		}
	}

	engine := service.NewAchievementService(service.NewBalanceService())
	earned, err := engine.Reevaluate(ctx, db, kid.ID)
	require.NoError(t, err)
	require.Len(t, earned, 1, "achievement earned at 3/3")

	// Void the first activity AFTER earning. The achievement must stay earned.
	reason := "later correction"
	require.NoError(t, q.VoidActivity(ctx, sqldb.VoidActivityParams{ID: first.ID, VoidReason: &reason}))

	// Re-eval to confirm — no revocation should happen.
	again, err := engine.Reevaluate(ctx, db, kid.ID)
	require.NoError(t, err)
	assert.Empty(t, again, "re-eval after void should not produce new earns")

	earnedList, err := q.ListEarnedAchievementsForKid(ctx, kid.ID)
	require.NoError(t, err)
	require.Len(t, earnedList, 1, "the original earned row must remain")
}

func TestAchievement_Reevaluate_CascadeRegardlessOfOrder(t *testing.T) {
	// B depends on A's bonus_points. B is created FIRST so it has the lower id
	// and would be evaluated before A in any naïve ordering. A single-pass
	// evaluator would miss B; fixed-point iteration earns both.
	ctx := context.Background()
	db := testutil.NewDB(t)
	q := sqldb.New(db)

	kid := testutil.NewKid(t, ctx, q, "Mia")
	chores := testutil.NewCategory(t, ctx, q, "chores", "Chores")
	dishes := testutil.NewActivityType(t, ctx, q, chores.ID, "wash-dishes", "Wash Dishes", 10, 5)

	// B: 50 global points — depends on A's bonus to pass.
	testutil.NewAchievement(t, ctx, q, testutil.AchievementSpec{
		Slug: "saver", Name: "Saver", Combinator: "ALL", BonusPoints: 0,
		Rules: []testutil.RuleSpec{{CategoryID: nil, Metric: "points", Threshold: 50}},
	})
	// A: 3 dish counts; grants 40 bonus points.
	testutil.NewAchievement(t, ctx, q, testutil.AchievementSpec{
		Slug: "dish-champ", Name: "Dish Champion", Combinator: "ALL", BonusPoints: 40,
		Rules: []testutil.RuleSpec{{CategoryID: &chores.ID, Metric: "count", Threshold: 3}},
	})

	// 3 dishes = 15 activity points; +40 bonus from A = 55, crosses B's 50 threshold.
	for i := 0; i < 3; i++ {
		testutil.LogActivity(t, ctx, q, kid.ID, dishes.ID, 1)
	}

	engine := service.NewAchievementService(service.NewBalanceService())
	earned, err := engine.Reevaluate(ctx, db, kid.ID)
	require.NoError(t, err)

	names := make([]string, 0, len(earned))
	for _, a := range earned {
		names = append(names, a.Name)
	}
	assert.ElementsMatch(t, []string{"Saver", "Dish Champion"}, names,
		"cascade: A grants points that satisfy B's rule — both earn in same call regardless of iteration order")
}

func TestAchievement_Reevaluate_GlobalPointsRuleSumsAllSources(t *testing.T) {
	ctx := context.Background()
	db := testutil.NewDB(t)
	q := sqldb.New(db)

	kid := testutil.NewKid(t, ctx, q, "Mia")
	chores := testutil.NewCategory(t, ctx, q, "chores", "Chores")
	dishes := testutil.NewActivityType(t, ctx, q, chores.ID, "wash-dishes", "Wash Dishes", 10, 5)
	testutil.NewAchievement(t, ctx, q, testutil.AchievementSpec{
		Slug: "saver", Name: "Saver", Combinator: "ALL",
		Rules: []testutil.RuleSpec{
			{CategoryID: nil, Metric: "points", Threshold: 100},
		},
	})

	// 5 dishes = 25 pts; plus a 75-pt adjustment = 100 pts total.
	for i := 0; i < 5; i++ {
		testutil.LogActivity(t, ctx, q, kid.ID, dishes.ID, 1)
	}
	reason := "great week"
	_, err := q.InsertPointAdjustment(ctx, sqldb.InsertPointAdjustmentParams{
		KidID: kid.ID, PointsDelta: 75, Reason: &reason,
	})
	require.NoError(t, err)

	engine := service.NewAchievementService(service.NewBalanceService())
	earned, err := engine.Reevaluate(ctx, db, kid.ID)
	require.NoError(t, err)
	require.Len(t, earned, 1)
}

func TestAchievement_Reevaluate_CategoryPointsRuleIgnoresOtherSources(t *testing.T) {
	ctx := context.Background()
	db := testutil.NewDB(t)
	q := sqldb.New(db)

	kid := testutil.NewKid(t, ctx, q, "Mia")
	chores := testutil.NewCategory(t, ctx, q, "chores", "Chores")
	dishes := testutil.NewActivityType(t, ctx, q, chores.ID, "wash-dishes", "Wash Dishes", 10, 5)
	testutil.NewAchievement(t, ctx, q, testutil.AchievementSpec{
		Slug: "chore-saver", Name: "Chore Saver", Combinator: "ALL",
		Rules: []testutil.RuleSpec{
			{CategoryID: &chores.ID, Metric: "points", Threshold: 50},
		},
	})

	// 5 dishes in chores = 25 pts in this category. Adjustment of 1000 pts must NOT count.
	for i := 0; i < 5; i++ {
		testutil.LogActivity(t, ctx, q, kid.ID, dishes.ID, 1)
	}
	reason := "irrelevant bonus"
	_, err := q.InsertPointAdjustment(ctx, sqldb.InsertPointAdjustmentParams{
		KidID: kid.ID, PointsDelta: 1000, Reason: &reason,
	})
	require.NoError(t, err)

	engine := service.NewAchievementService(service.NewBalanceService())
	earned, err := engine.Reevaluate(ctx, db, kid.ID)
	require.NoError(t, err)
	assert.Empty(t, earned, "category-scoped points must ignore non-activity sources")

	// Now log 5 more dishes — 50 pts in category, should pass.
	for i := 0; i < 5; i++ {
		testutil.LogActivity(t, ctx, q, kid.ID, dishes.ID, 1)
	}
	earned, err = engine.Reevaluate(ctx, db, kid.ID)
	require.NoError(t, err)
	require.Len(t, earned, 1)
}

func TestAchievement_Reevaluate_LevelRulePassesWhenXPCrossesThreshold(t *testing.T) {
	ctx := context.Background()
	db := testutil.NewDB(t)
	q := sqldb.New(db)

	kid := testutil.NewKid(t, ctx, q, "Mia")
	chores := testutil.NewCategory(t, ctx, q, "chores", "Chores")
	dishes := testutil.NewActivityType(t, ctx, q, chores.ID, "wash-dishes", "Wash Dishes", 10, 5)
	// Level 3 == 300 XP per the curve.
	testutil.NewAchievement(t, ctx, q, testutil.AchievementSpec{
		Slug: "tier3", Name: "Tier 3 Dishwasher", Combinator: "ALL",
		Rules: []testutil.RuleSpec{
			{CategoryID: &chores.ID, Metric: "level", Threshold: 3},
		},
	})

	// 30 dishes × 10 XP = 300 XP exactly.
	for i := 0; i < 30; i++ {
		testutil.LogActivity(t, ctx, q, kid.ID, dishes.ID, 1)
	}

	engine := service.NewAchievementService(service.NewBalanceService())
	earned, err := engine.Reevaluate(ctx, db, kid.ID)
	require.NoError(t, err)
	require.Len(t, earned, 1)
}

func TestAchievement_Reevaluate_LevelRuleBelowThresholdDoesNotPass(t *testing.T) {
	ctx := context.Background()
	db := testutil.NewDB(t)
	q := sqldb.New(db)

	kid := testutil.NewKid(t, ctx, q, "Mia")
	chores := testutil.NewCategory(t, ctx, q, "chores", "Chores")
	dishes := testutil.NewActivityType(t, ctx, q, chores.ID, "wash-dishes", "Wash Dishes", 10, 5)
	testutil.NewAchievement(t, ctx, q, testutil.AchievementSpec{
		Slug: "tier3", Name: "Tier 3 Dishwasher", Combinator: "ALL",
		Rules: []testutil.RuleSpec{
			{CategoryID: &chores.ID, Metric: "level", Threshold: 3},
		},
	})

	// 29 dishes × 10 XP = 290 XP — one short of 300.
	for i := 0; i < 29; i++ {
		testutil.LogActivity(t, ctx, q, kid.ID, dishes.ID, 1)
	}

	engine := service.NewAchievementService(service.NewBalanceService())
	earned, err := engine.Reevaluate(ctx, db, kid.ID)
	require.NoError(t, err)
	assert.Empty(t, earned)
}

func TestAchievement_Reevaluate_XPRulePasses(t *testing.T) {
	ctx := context.Background()
	db := testutil.NewDB(t)
	q := sqldb.New(db)

	kid := testutil.NewKid(t, ctx, q, "Mia")
	chores := testutil.NewCategory(t, ctx, q, "chores", "Chores")
	dishes := testutil.NewActivityType(t, ctx, q, chores.ID, "wash-dishes", "Wash Dishes", 10, 5)
	testutil.NewAchievement(t, ctx, q, testutil.AchievementSpec{
		Slug: "xp-rookie", Name: "XP Rookie", Combinator: "ALL",
		Rules: []testutil.RuleSpec{
			{CategoryID: &chores.ID, Metric: "xp", Threshold: 30},
		},
	})

	for i := 0; i < 3; i++ {
		testutil.LogActivity(t, ctx, q, kid.ID, dishes.ID, 1) // 10 XP each → 30 total
	}

	engine := service.NewAchievementService(service.NewBalanceService())
	earned, err := engine.Reevaluate(ctx, db, kid.ID)
	require.NoError(t, err)
	require.Len(t, earned, 1)
}

func TestAchievement_Reevaluate_CountRulePasses(t *testing.T) {
	ctx := context.Background()
	db := testutil.NewDB(t)
	q := sqldb.New(db)

	kid := testutil.NewKid(t, ctx, q, "Mia")
	chores := testutil.NewCategory(t, ctx, q, "chores", "Chores")
	dishes := testutil.NewActivityType(t, ctx, q, chores.ID, "wash-dishes", "Wash Dishes", 10, 5)

	ach := testutil.NewAchievement(t, ctx, q, testutil.AchievementSpec{
		Slug: "dish-champ", Name: "Dish Champion", Combinator: "ALL", BonusPoints: 50,
		Rules: []testutil.RuleSpec{
			{CategoryID: &chores.ID, Metric: "count", Threshold: 3},
		},
	})

	for i := 0; i < 3; i++ {
		testutil.LogActivity(t, ctx, q, kid.ID, dishes.ID, 1)
	}

	engine := service.NewAchievementService(service.NewBalanceService())
	earned, err := engine.Reevaluate(ctx, db, kid.ID)
	require.NoError(t, err)
	require.Len(t, earned, 1)
	assert.Equal(t, ach.ID, earned[0].ID)
	assert.Equal(t, "Dish Champion", earned[0].Name)
}
