package domain

import "time"

type Kid struct {
	ID           int64
	Name         string
	AvatarSlug   string
	Color        string
	DisplayOrder int
	ArchivedAt   *time.Time
	CreatedAt    time.Time
}
