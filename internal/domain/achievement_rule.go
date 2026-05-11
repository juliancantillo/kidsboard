package domain

type Metric string

const (
	MetricCount  Metric = "count"
	MetricXP     Metric = "xp"
	MetricPoints Metric = "points"
	MetricLevel  Metric = "level"
)

type AchievementRule struct {
	ID            int64
	AchievementID int64
	CategoryID    *int64
	Metric        Metric
	Threshold     int64
}
