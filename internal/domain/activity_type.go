package domain

import "time"

type ActivityType struct {
	ID            int64
	CategoryID    int64
	Slug          string
	Name          string
	Description   *string
	XPPerUnit     int64
	PointsPerUnit int64
	ArchivedAt    *time.Time
	CreatedAt     time.Time
}
