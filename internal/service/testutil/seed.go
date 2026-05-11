package testutil

import (
	"context"
	"testing"

	"cantillo.dev/kidsboard/internal/storage/sqldb"
	"github.com/stretchr/testify/require"
)

// NewKid inserts a kid with sensible defaults and returns the created row.
// Pass-through helpers like this keep test setup terse.
func NewKid(t *testing.T, ctx context.Context, q *sqldb.Queries, name string) sqldb.Kid {
	t.Helper()
	kid, err := q.CreateKid(ctx, sqldb.CreateKidParams{
		Name: name, AvatarSlug: "mage", Color: "#6366F1", DisplayOrder: 0,
	})
	require.NoError(t, err)
	return kid
}

func NewCategory(t *testing.T, ctx context.Context, q *sqldb.Queries, slug, name string) sqldb.Category {
	t.Helper()
	cat, err := q.CreateCategory(ctx, sqldb.CreateCategoryParams{
		Slug: slug, Name: name,
	})
	require.NoError(t, err)
	return cat
}

func NewActivityType(t *testing.T, ctx context.Context, q *sqldb.Queries, categoryID int64, slug, name string, xpPerUnit, pointsPerUnit int64) sqldb.ActivityType {
	t.Helper()
	at, err := q.CreateActivityType(ctx, sqldb.CreateActivityTypeParams{
		CategoryID: categoryID, Slug: slug, Name: name,
		XpPerUnit: xpPerUnit, PointsPerUnit: pointsPerUnit,
	})
	require.NoError(t, err)
	return at
}

// AchievementSpec is a parent + rules pair the test wants inserted together.
type AchievementSpec struct {
	Slug        string
	Name        string
	Combinator  string // "ALL" or "ANY"
	BonusPoints int64
	Rules       []RuleSpec
}

type RuleSpec struct {
	CategoryID *int64 // nil = global
	Metric     string // "count" | "xp" | "points" | "level"
	Threshold  int64
}

func NewAchievement(t *testing.T, ctx context.Context, q *sqldb.Queries, spec AchievementSpec) sqldb.Achievement {
	t.Helper()
	ach, err := q.CreateAchievement(ctx, sqldb.CreateAchievementParams{
		Slug:        spec.Slug,
		Name:        spec.Name,
		Combinator:  spec.Combinator,
		BonusPoints: spec.BonusPoints,
	})
	require.NoError(t, err)
	for _, r := range spec.Rules {
		_, err := q.InsertAchievementRule(ctx, sqldb.InsertAchievementRuleParams{
			AchievementID: ach.ID,
			CategoryID:    r.CategoryID,
			Metric:        r.Metric,
			Threshold:     r.Threshold,
		})
		require.NoError(t, err)
	}
	return ach
}

// LogActivity inserts an activity with the given quantity, snapshotting reward totals
// from the activity_type. Mirrors what the future ActivityService.Log will do internally.
func LogActivity(t *testing.T, ctx context.Context, q *sqldb.Queries, kidID, activityTypeID int64, quantity int) sqldb.Activity {
	t.Helper()
	at, err := q.GetActivityType(ctx, activityTypeID)
	require.NoError(t, err)
	act, err := q.InsertActivity(ctx, sqldb.InsertActivityParams{
		KidID:          kidID,
		ActivityTypeID: activityTypeID,
		Quantity:       int64(quantity),
		XpAwarded:      at.XpPerUnit * int64(quantity),
		PointsAwarded:  at.PointsPerUnit * int64(quantity),
	})
	require.NoError(t, err)
	return act
}
