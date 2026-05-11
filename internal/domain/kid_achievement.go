package domain

import "time"

type KidAchievement struct {
	KidID         int64
	AchievementID int64
	EarnedAt      time.Time
	Unseen        bool
}
