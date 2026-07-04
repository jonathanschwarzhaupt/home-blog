package models

import (
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

var (
	ErrNoRecord      = errors.New("models: no matching record found")
	ErrEditConflict  = errors.New("models: edit conflict")
	ErrDuplicateSlug = errors.New("models: a post with this slug already exists")
)

const pgUniqueViolationCode = "23505"

// WrapDBError translates Postgres/pgx-specific errors into models sentinel
// errors, so callers check against these instead of driver-specific types.
func WrapDBError(err error) error {
	if err == nil {
		return nil
	}

	if errors.Is(err, pgx.ErrNoRows) {
		return ErrNoRecord
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == pgUniqueViolationCode {
		return ErrDuplicateSlug
	}

	return err
}
