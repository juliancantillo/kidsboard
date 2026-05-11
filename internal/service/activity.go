package service

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"cantillo.dev/kidsboard/internal/domain"
	"cantillo.dev/kidsboard/internal/storage"
	"cantillo.dev/kidsboard/internal/storage/sqldb"
)

// ActivityService logs and voids activities. Each log opens a transaction,
// inserts the activity row with snapshotted rewards, and re-evaluates
// achievements in the same tx — so the kid sees any newly-earned achievement
// in the same response cycle.
type ActivityService interface {
	Log(ctx context.Context, db *sql.DB, in LogActivityInput) (LogResult, error)
}

type LogActivityInput struct {
	KidID          int64
	ActivityTypeID int64
	Quantity       int
	Note           string
	OccurredAt     *time.Time
}

// LogResult bundles the inserted activity with any achievements that earned
// in the same transaction. Useful for surfacing celebration moments.
type LogResult struct {
	Activity    domain.Activity
	NewlyEarned []domain.Achievement
}

func NewActivityService(activityTypes ActivityTypeService, achievements AchievementService) ActivityService {
	return &activityService{
		activityTypes: activityTypes,
		achievements:  achievements,
	}
}

type activityService struct {
	activityTypes ActivityTypeService
	achievements  AchievementService
}

func (s *activityService) Log(ctx context.Context, db *sql.DB, in LogActivityInput) (LogResult, error) {
	if in.Quantity < 1 {
		in.Quantity = 1
	}
	return storage.WithTx(ctx, db, func(tx storage.DBTX) (LogResult, error) {
		at, err := s.activityTypes.Get(ctx, tx, in.ActivityTypeID)
		if err != nil {
			return LogResult{}, err
		}
		if at.ArchivedAt != nil {
			return LogResult{}, fmt.Errorf("activity type archived: %w", ErrInvalidInput)
		}

		note := strings.TrimSpace(in.Note)
		var notePtr *string
		if note != "" {
			notePtr = &note
		}
		params := sqldb.InsertActivityParams{
			KidID:          in.KidID,
			ActivityTypeID: at.ID,
			Quantity:       int64(in.Quantity),
			XpAwarded:      at.XPPerUnit * int64(in.Quantity),
			PointsAwarded:  at.PointsPerUnit * int64(in.Quantity),
			Note:           notePtr,
		}
		if in.OccurredAt != nil {
			params.OccurredAt = *in.OccurredAt
		} else {
			params.OccurredAt = time.Now()
		}

		row, err := sqldb.New(tx).InsertActivity(ctx, params)
		if err != nil {
			return LogResult{}, fmt.Errorf("insert activity: %w", err)
		}

		earned, err := s.achievements.Reevaluate(ctx, tx, in.KidID)
		if err != nil {
			return LogResult{}, err
		}

		return LogResult{
			Activity: domain.Activity{
				ID: row.ID, KidID: row.KidID, ActivityTypeID: row.ActivityTypeID,
				Quantity:  int(row.Quantity),
				XPAwarded: row.XpAwarded, PointsAwarded: row.PointsAwarded,
				Note: row.Note, OccurredAt: row.OccurredAt, CreatedAt: row.CreatedAt,
				VoidedAt: row.VoidedAt, VoidReason: row.VoidReason,
			},
			NewlyEarned: earned,
		}, nil
	})
}
