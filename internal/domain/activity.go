package domain

import "time"

type Activity struct {
	ID             int64
	KidID          int64
	ActivityTypeID int64
	Quantity       int
	XPAwarded      int64
	PointsAwarded  int64
	Note           *string
	OccurredAt     time.Time
	CreatedAt      time.Time
	VoidedAt       *time.Time
	VoidReason     *string
}
