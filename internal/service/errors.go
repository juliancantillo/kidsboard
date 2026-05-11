package service

import (
	"database/sql"
	"errors"
)

// Domain-level sentinel errors. Repositories translate sql.ErrNoRows into
// ErrNotFound at the boundary; controllers map sentinels to HTTP statuses.
var (
	ErrNotFound                 = errors.New("not found")
	ErrAlreadyExists            = errors.New("already exists")
	ErrInsufficientBalance      = errors.New("insufficient balance")
	ErrAchievementAlreadyEarned = errors.New("achievement already earned")
	ErrRedemptionAlreadyDecided = errors.New("redemption already decided")
	ErrRewardInactive           = errors.New("reward not active")
	ErrInvalidInput             = errors.New("invalid input")
)

// ValidationError carries per-field error messages for form submissions.
// Returned by service methods alongside (or in place of) the sentinels above
// when DTO validation fails. Unwrap to ErrInvalidInput so callers can switch
// generically.
type ValidationError struct {
	Fields map[string]string
}

func (e *ValidationError) Error() string { return "validation failed" }
func (e *ValidationError) Unwrap() error { return ErrInvalidInput }

// errNoRows returns the sentinel database/sql uses for empty result sets.
// Wrapped as a function so callers can do `errors.Is(err, errNoRows())`
// without importing database/sql directly.
func errNoRows() error { return sql.ErrNoRows }
