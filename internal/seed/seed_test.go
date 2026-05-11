package seed_test

import (
	"context"
	"testing"

	"cantillo.dev/kidsboard/internal/seed"
	"cantillo.dev/kidsboard/internal/service/testutil"
	"cantillo.dev/kidsboard/internal/storage/sqldb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSeed_PopulatesAllCuratedConfig(t *testing.T) {
	ctx := context.Background()
	db := testutil.NewDB(t)

	require.NoError(t, seed.Run(ctx, db))

	stats, err := seed.Count(ctx, db)
	require.NoError(t, err)
	assert.Equal(t, len(seed.Categories), stats.Categories)
	assert.Equal(t, len(seed.ActivityTypes), stats.ActivityTypes)
	assert.Equal(t, len(seed.Achievements), stats.Achievements)

	expectedRules := 0
	for _, a := range seed.Achievements {
		expectedRules += len(a.Rules)
	}
	assert.Equal(t, expectedRules, stats.AchievementRules)
}

func TestSeed_IsIdempotent(t *testing.T) {
	ctx := context.Background()
	db := testutil.NewDB(t)

	require.NoError(t, seed.Run(ctx, db))
	first, err := seed.Count(ctx, db)
	require.NoError(t, err)

	require.NoError(t, seed.Run(ctx, db))
	second, err := seed.Count(ctx, db)
	require.NoError(t, err)

	assert.Equal(t, first, second, "re-running seed must not change counts")
}

func TestSeed_PreservesIDsAcrossRuns(t *testing.T) {
	ctx := context.Background()
	db := testutil.NewDB(t)
	q := sqldb.New(db)

	require.NoError(t, seed.Run(ctx, db))
	firstAch, err := q.GetAchievementBySlug(ctx, "quehaceres-antman")
	require.NoError(t, err)

	require.NoError(t, seed.Run(ctx, db))
	secondAch, err := q.GetAchievementBySlug(ctx, "quehaceres-antman")
	require.NoError(t, err)

	assert.Equal(t, firstAch.ID, secondAch.ID,
		"upsert by slug must keep the same ID — kid_achievements FKs depend on it")
}

func TestSeed_AntmanAchievementHasRightFlavor(t *testing.T) {
	ctx := context.Background()
	db := testutil.NewDB(t)
	q := sqldb.New(db)
	require.NoError(t, seed.Run(ctx, db))

	ach, err := q.GetAchievementBySlug(ctx, "quehaceres-antman")
	require.NoError(t, err)
	assert.Equal(t, "Antman del Hogar", ach.Name)
	require.NotNil(t, ach.Title)
	assert.Contains(t, *ach.Title, "Nanoverso")
	rules, err := q.ListAchievementRules(ctx, ach.ID)
	require.NoError(t, err)
	require.Len(t, rules, 1)
	assert.Equal(t, "count", rules[0].Metric)
	assert.Equal(t, int64(200), rules[0].Threshold)
}
