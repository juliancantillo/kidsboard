package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"cantillo.dev/kidsboard/internal/domain"
	"cantillo.dev/kidsboard/internal/storage"
	"cantillo.dev/kidsboard/internal/storage/sqldb"
)

// ActivityTypeService is the read API for activity types.
type ActivityTypeService interface {
	Get(ctx context.Context, db storage.DBTX, id int64) (domain.ActivityType, error)
	ListActive(ctx context.Context, db storage.DBTX) ([]domain.ActivityType, error)
	ListByCategory(ctx context.Context, db storage.DBTX, categoryID int64) ([]domain.ActivityType, error)
}

func NewActivityTypeService() ActivityTypeService { return &activityTypeService{} }

type activityTypeService struct{}

func (s *activityTypeService) Get(ctx context.Context, db storage.DBTX, id int64) (domain.ActivityType, error) {
	row, err := sqldb.New(db).GetActivityType(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.ActivityType{}, ErrNotFound
		}
		return domain.ActivityType{}, fmt.Errorf("get activity type: %w", err)
	}
	return activityTypeFromRow(row), nil
}

func (s *activityTypeService) ListActive(ctx context.Context, db storage.DBTX) ([]domain.ActivityType, error) {
	rows, err := sqldb.New(db).ListActivityTypes(ctx)
	if err != nil {
		return nil, fmt.Errorf("list activity types: %w", err)
	}
	out := make([]domain.ActivityType, 0, len(rows))
	for _, r := range rows {
		out = append(out, activityTypeFromRow(r))
	}
	return out, nil
}

func (s *activityTypeService) ListByCategory(ctx context.Context, db storage.DBTX, categoryID int64) ([]domain.ActivityType, error) {
	rows, err := sqldb.New(db).ListActivityTypesByCategory(ctx, categoryID)
	if err != nil {
		return nil, fmt.Errorf("list activity types by category: %w", err)
	}
	out := make([]domain.ActivityType, 0, len(rows))
	for _, r := range rows {
		out = append(out, activityTypeFromRow(r))
	}
	return out, nil
}

func activityTypeFromRow(r sqldb.ActivityType) domain.ActivityType {
	return domain.ActivityType{
		ID: r.ID, CategoryID: r.CategoryID, Slug: r.Slug, Name: r.Name,
		Description: r.Description,
		XPPerUnit: r.XpPerUnit, PointsPerUnit: r.PointsPerUnit,
		ArchivedAt: r.ArchivedAt, CreatedAt: r.CreatedAt,
	}
}
