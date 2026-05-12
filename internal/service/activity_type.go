package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"cantillo.dev/kidsboard/internal/domain"
	"cantillo.dev/kidsboard/internal/storage"
	"cantillo.dev/kidsboard/internal/storage/sqldb"
)

// ActivityTypeService is the parent-facing CRUD surface for activity types —
// the per-category templates (with XP + points config) that activities are
// logged against.
type ActivityTypeService interface {
	Get(ctx context.Context, db storage.DBTX, id int64) (domain.ActivityType, error)
	ListActive(ctx context.Context, db storage.DBTX) ([]domain.ActivityType, error)
	ListAll(ctx context.Context, db storage.DBTX) ([]domain.ActivityType, error)
	ListByCategory(ctx context.Context, db storage.DBTX, categoryID int64) ([]domain.ActivityType, error)

	Create(ctx context.Context, db storage.DBTX, in ActivityTypeInput) (domain.ActivityType, error)
	Update(ctx context.Context, db storage.DBTX, id int64, in ActivityTypeInput) (domain.ActivityType, error)
	Archive(ctx context.Context, db storage.DBTX, id int64) error
	Unarchive(ctx context.Context, db storage.DBTX, id int64) error
}

// ActivityTypeInput is the parent-facing form shape for create/update.
// Description empty-string == NULL in DB.
type ActivityTypeInput struct {
	CategoryID    int64
	Slug          string
	Name          string
	Description   string
	XPPerUnit     int64
	PointsPerUnit int64
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

func (s *activityTypeService) ListAll(ctx context.Context, db storage.DBTX) ([]domain.ActivityType, error) {
	rows, err := sqldb.New(db).ListAllActivityTypes(ctx)
	if err != nil {
		return nil, fmt.Errorf("list all activity types: %w", err)
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

func (s *activityTypeService) Create(ctx context.Context, db storage.DBTX, in ActivityTypeInput) (domain.ActivityType, error) {
	if err := s.validate(in); err != nil {
		return domain.ActivityType{}, err
	}
	row, err := sqldb.New(db).CreateActivityType(ctx, sqldb.CreateActivityTypeParams{
		CategoryID:    in.CategoryID,
		Slug:          in.Slug,
		Name:          strings.TrimSpace(in.Name),
		Description:   optString(in.Description),
		XpPerUnit:     in.XPPerUnit,
		PointsPerUnit: in.PointsPerUnit,
	})
	if err != nil {
		if isUniqueConstraint(err) {
			return domain.ActivityType{}, &ValidationError{Fields: map[string]string{
				"slug": "Ya existe una actividad con ese slug.",
			}}
		}
		return domain.ActivityType{}, fmt.Errorf("create activity type: %w", err)
	}
	return activityTypeFromRow(row), nil
}

func (s *activityTypeService) Update(ctx context.Context, db storage.DBTX, id int64, in ActivityTypeInput) (domain.ActivityType, error) {
	if err := s.validate(in); err != nil {
		return domain.ActivityType{}, err
	}
	q := sqldb.New(db)
	if err := q.UpdateActivityType(ctx, sqldb.UpdateActivityTypeParams{
		ID:            id,
		CategoryID:    in.CategoryID,
		Name:          strings.TrimSpace(in.Name),
		Description:   optString(in.Description),
		XpPerUnit:     in.XPPerUnit,
		PointsPerUnit: in.PointsPerUnit,
	}); err != nil {
		return domain.ActivityType{}, fmt.Errorf("update activity type: %w", err)
	}
	row, err := q.GetActivityType(ctx, id)
	if err != nil {
		return domain.ActivityType{}, fmt.Errorf("reload activity type: %w", err)
	}
	return activityTypeFromRow(row), nil
}

func (s *activityTypeService) Archive(ctx context.Context, db storage.DBTX, id int64) error {
	return sqldb.New(db).ArchiveActivityType(ctx, id)
}

func (s *activityTypeService) Unarchive(ctx context.Context, db storage.DBTX, id int64) error {
	return sqldb.New(db).UnarchiveActivityType(ctx, id)
}

// -- Validation -------------------------------------------------------------

func (s *activityTypeService) validate(in ActivityTypeInput) error {
	fields := map[string]string{}
	if in.CategoryID <= 0 {
		fields["category_id"] = "Selecciona una categoría."
	}
	if !isValidSlug(in.Slug) {
		fields["slug"] = "Slug requerido (a-z, 0-9, guiones; 1-60 caracteres)."
	}
	if name := strings.TrimSpace(in.Name); name == "" || len(name) > 100 {
		fields["name"] = "El nombre debe tener entre 1 y 100 caracteres."
	}
	if in.XPPerUnit < 0 {
		fields["xp_per_unit"] = "El XP no puede ser negativo."
	}
	if in.PointsPerUnit < 0 {
		fields["points_per_unit"] = "Los puntos no pueden ser negativos."
	}
	if len(fields) > 0 {
		return &ValidationError{Fields: fields}
	}
	return nil
}

func isUniqueConstraint(err error) bool {
	return err != nil && strings.Contains(err.Error(), "UNIQUE constraint failed")
}

func activityTypeFromRow(r sqldb.ActivityType) domain.ActivityType {
	return domain.ActivityType{
		ID: r.ID, CategoryID: r.CategoryID, Slug: r.Slug, Name: r.Name,
		Description: r.Description,
		XPPerUnit:   r.XpPerUnit, PointsPerUnit: r.PointsPerUnit,
		ArchivedAt: r.ArchivedAt, CreatedAt: r.CreatedAt,
	}
}
