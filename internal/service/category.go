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

// CategoryService is the read API for categories (seeded via `kidsboard seed`).
type CategoryService interface {
	Get(ctx context.Context, db storage.DBTX, id int64) (domain.Category, error)
	ListActive(ctx context.Context, db storage.DBTX) ([]domain.Category, error)
}

func NewCategoryService() CategoryService { return &categoryService{} }

type categoryService struct{}

func (s *categoryService) Get(ctx context.Context, db storage.DBTX, id int64) (domain.Category, error) {
	row, err := sqldb.New(db).GetCategory(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Category{}, ErrNotFound
		}
		return domain.Category{}, fmt.Errorf("get category: %w", err)
	}
	return categoryFromRow(row), nil
}

func (s *categoryService) ListActive(ctx context.Context, db storage.DBTX) ([]domain.Category, error) {
	rows, err := sqldb.New(db).ListCategories(ctx)
	if err != nil {
		return nil, fmt.Errorf("list categories: %w", err)
	}
	out := make([]domain.Category, 0, len(rows))
	for _, r := range rows {
		out = append(out, categoryFromRow(r))
	}
	return out, nil
}

func categoryFromRow(r sqldb.Category) domain.Category {
	return domain.Category{
		ID: r.ID, Slug: r.Slug, Name: r.Name,
		Description: r.Description, Icon: r.Icon, Color: r.Color,
		ArchivedAt: r.ArchivedAt, CreatedAt: r.CreatedAt,
	}
}
