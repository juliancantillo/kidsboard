package domain

import "time"

type Category struct {
	ID          int64
	Slug        string
	Name        string
	Description *string
	Icon        *string
	Color       *string
	ArchivedAt  *time.Time
	CreatedAt   time.Time
}
