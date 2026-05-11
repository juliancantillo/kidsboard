package domain

import "time"

type RedemptionStatus string

const (
	RedemptionPending   RedemptionStatus = "pending"
	RedemptionApproved  RedemptionStatus = "approved"
	RedemptionRejected  RedemptionStatus = "rejected"
	RedemptionCancelled RedemptionStatus = "cancelled"
)

type Redemption struct {
	ID          int64
	KidID       int64
	RewardID    int64
	PointsSpent int64
	Status      RedemptionStatus
	RequestedAt time.Time
	DecidedAt   *time.Time
}
