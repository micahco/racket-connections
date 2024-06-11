package models

import (
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
)

var (
	ErrNoRecord           = errors.New("models: no matching record found")
	ErrInvalidCredentials = errors.New("models: invalid credentials")
	ErrNotVerified        = errors.New("models: user not verified")
	ErrDuplicateEmail     = errors.New("models: duplicate email")
)

func pgErrCode(err error) string {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code
	}

	return ""
}
