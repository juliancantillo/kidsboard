package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLevelForXP_ZeroXPIsLevelOne(t *testing.T) {
	assert.Equal(t, 1, LevelForXP(0))
}

func TestLevelForXP_BelowNextThresholdStays(t *testing.T) {
	// 250 is between L2 (100) and L3 (300) — kid is still Level 2.
	assert.Equal(t, 2, LevelForXP(250))
}

func TestLevelForXP_AtExactThresholdAdvances(t *testing.T) {
	// 300 is exactly L3's threshold — kid is Level 3, not 2.
	assert.Equal(t, 3, LevelForXP(300))
}

func TestLevelForXP_RoundTrip(t *testing.T) {
	for n := 1; n <= 30; n++ {
		assert.Equalf(t, n, LevelForXP(XPForLevel(n)), "round-trip at level %d", n)
	}
}

func TestLevelForXP_Monotonic(t *testing.T) {
	prev := LevelForXP(0)
	for xp := int64(1); xp <= 20000; xp += 37 {
		current := LevelForXP(xp)
		assert.GreaterOrEqualf(t, current, prev, "level dropped between xp=%d and xp=%d", xp-37, xp)
		prev = current
	}
}

func TestXPForNextLevel_GivesNextThreshold(t *testing.T) {
	// Kid at 150 XP is Level 2; next threshold is L3 = 300.
	assert.Equal(t, int64(300), XPForNextLevel(150))
	// Kid at exactly L3 (300 XP) — next is L4 = 600.
	assert.Equal(t, int64(600), XPForNextLevel(300))
}

func TestXPForLevel_MatchesCurve(t *testing.T) {
	cases := []struct {
		level int
		xp    int64
	}{
		{1, 0},
		{2, 100},
		{3, 300},
		{4, 600},
		{5, 1000},
		{10, 4500},
		{20, 19000},
	}
	for _, c := range cases {
		assert.Equalf(t, c.xp, XPForLevel(c.level), "XPForLevel(%d)", c.level)
	}
}
