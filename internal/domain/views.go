package domain

import "time"

// CategoryLevel describes a kid's progression within a single category.
type CategoryLevel struct {
	Category       Category
	XP             int64
	Level          int
	XPIntoLevel    int64
	XPForNext      int64
	ProgressRatio  float64
}

// RuleProgress carries the live values for a single rule of an achievement.
type RuleProgress struct {
	Rule      AchievementRule
	Current   int64
	Threshold int64
	Ratio     float64
}

// AchievementProgress is an unearned achievement with its current per-rule
// state and an overall ratio that respects the achievement's combinator.
type AchievementProgress struct {
	Achievement  Achievement
	Rules        []RuleProgress
	OverallRatio float64
}

// Balance bundles a kid's points across the four sources the system tracks.
type Balance struct {
	Earned           int64
	Spent            int64
	Reserved         int64
	Balance          int64
	AvailableBalance int64
}

// EarnedAchievement pairs an earned row with the full achievement definition.
type EarnedAchievement struct {
	Achievement Achievement
	EarnedAt    time.Time
	Unseen      bool
}

// ProfileView is the composite read model rendered on a kid's profile page.
type ProfileView struct {
	Kid             Kid
	Balance         Balance
	CategoryLevels  []CategoryLevel
	NextAchievements []AchievementProgress
	RecentEarned    []EarnedAchievement
	RecentActivity  []Activity
	PendingRedemptions []Redemption
}
