package domain

import "time"

type Combinator string

const (
	CombinatorAll Combinator = "ALL"
	CombinatorAny Combinator = "ANY"
)

type Achievement struct {
	ID          int64
	Slug        string
	Name        string
	Description *string
	Title       *string
	Combinator  Combinator
	BonusPoints int64
	Rules       []AchievementRule
	ArchivedAt  *time.Time
	CreatedAt   time.Time
}
