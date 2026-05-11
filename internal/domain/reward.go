package domain

import "time"

type Reward struct {
	ID          int64
	Slug        string
	Name        string
	Description *string
	CostPoints  int64
	Active      bool
	ArchivedAt  *time.Time
	CreatedAt   time.Time
}
