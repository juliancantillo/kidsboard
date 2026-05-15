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

// KidService owns kid CRUD and listing. Controllers must use this rather
// than touching sqldb directly — keeps the HTTP layer ignorant of storage.
type KidService interface {
	Get(ctx context.Context, db storage.DBTX, id int64) (domain.Kid, error)
	ListActive(ctx context.Context, db storage.DBTX) ([]domain.Kid, error)
	ListAll(ctx context.Context, db storage.DBTX) ([]domain.Kid, error)
	Create(ctx context.Context, db storage.DBTX, in CreateKidInput) (domain.Kid, error)
	Archive(ctx context.Context, db storage.DBTX, id int64) error
	Unarchive(ctx context.Context, db storage.DBTX, id int64) error
}

// CreateKidInput is the form-shape used by the admin parent flow.
// avatar_slug membership is validated against the static avatar whitelist
// passed at service construction.
type CreateKidInput struct {
	Name         string
	AvatarSlug   string
	Color        string
	DisplayOrder int
}

func NewKidService(avatarSlugs map[string]struct{}) KidService {
	return &kidService{avatars: avatarSlugs}
}

type kidService struct {
	avatars map[string]struct{}
}

func (s *kidService) Get(ctx context.Context, db storage.DBTX, id int64) (domain.Kid, error) {
	row, err := sqldb.New(db).GetKid(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Kid{}, ErrNotFound
		}
		return domain.Kid{}, fmt.Errorf("get kid: %w", err)
	}
	return kidFromRow(row), nil
}

func (s *kidService) ListActive(ctx context.Context, db storage.DBTX) ([]domain.Kid, error) {
	rows, err := sqldb.New(db).ListKids(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("list kids: %w", err)
	}
	out := make([]domain.Kid, 0, len(rows))
	for _, r := range rows {
		out = append(out, kidFromRow(r))
	}
	return out, nil
}

func (s *kidService) ListAll(ctx context.Context, db storage.DBTX) ([]domain.Kid, error) {
	rows, err := sqldb.New(db).ListKids(ctx, int64(1))
	if err != nil {
		return nil, fmt.Errorf("list kids: %w", err)
	}
	out := make([]domain.Kid, 0, len(rows))
	for _, r := range rows {
		out = append(out, kidFromRow(r))
	}
	return out, nil
}

func (s *kidService) Create(ctx context.Context, db storage.DBTX, in CreateKidInput) (domain.Kid, error) {
	if err := s.validateCreate(in); err != nil {
		return domain.Kid{}, err
	}
	row, err := sqldb.New(db).CreateKid(ctx, sqldb.CreateKidParams{
		Name:         in.Name,
		AvatarSlug:   in.AvatarSlug,
		Color:        in.Color,
		DisplayOrder: int64(in.DisplayOrder),
	})
	if err != nil {
		return domain.Kid{}, fmt.Errorf("create kid: %w", err)
	}
	return kidFromRow(row), nil
}

func (s *kidService) Archive(ctx context.Context, db storage.DBTX, id int64) error {
	if err := sqldb.New(db).ArchiveKid(ctx, id); err != nil {
		return fmt.Errorf("archive kid: %w", err)
	}
	return nil
}

func (s *kidService) Unarchive(ctx context.Context, db storage.DBTX, id int64) error {
	if err := sqldb.New(db).UnarchiveKid(ctx, id); err != nil {
		return fmt.Errorf("unarchive kid: %w", err)
	}
	return nil
}

func (s *kidService) validateCreate(in CreateKidInput) error {
	fields := map[string]string{}
	if name := strings.TrimSpace(in.Name); name == "" || len(name) > 50 {
		fields["name"] = "El nombre debe tener entre 1 y 50 caracteres."
	}
	if _, ok := s.avatars[in.AvatarSlug]; !ok {
		fields["avatar_slug"] = "Selecciona un avatar válido."
	}
	if !looksLikeHexColor(in.Color) {
		fields["color"] = "Color hex inválido (ej. #6366F1)."
	}
	if len(fields) > 0 {
		return &ValidationError{Fields: fields}
	}
	return nil
}

func kidFromRow(r sqldb.Kid) domain.Kid {
	return domain.Kid{
		ID:           r.ID,
		Name:         r.Name,
		AvatarSlug:   r.AvatarSlug,
		Color:        r.Color,
		DisplayOrder: int(r.DisplayOrder),
		ArchivedAt:   r.ArchivedAt,
		CreatedAt:    r.CreatedAt,
	}
}

func looksLikeHexColor(s string) bool {
	if len(s) != 7 || s[0] != '#' {
		return false
	}
	for _, c := range s[1:] {
		ok := (c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')
		if !ok {
			return false
		}
	}
	return true
}
