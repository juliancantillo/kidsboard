package domain

import "time"

type PointAdjustment struct {
	ID           int64
	KidID        int64
	PointsDelta  int64
	Reason       *string
	CreatedAt    time.Time
	VoidedAt     *time.Time
	VoidReason   *string
}
