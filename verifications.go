package main

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
)

type verification struct {
	email  string
	token  string
	expiry time.Time
}

func (v verification) isExpired() bool {
	return v.expiry.Before(time.Now())
}

func scanVerification(row pgx.CollectableRow) (*verification, error) {
	var v verification
	err := row.Scan(
		&v.email,
		&v.token,
		&v.expiry)
	return &v, err
}

func (db *pgStore) createVerification(email string, token string, expiry time.Time) (*verification, error) {
	sql := `INSERT INTO verifications
		(email, token, expiry)
		VALUES($1, $2, $3) RETURNING *;`
	rows, err := db.Query(context.Background(), sql,
		email, token, expiry)
	if err != nil {
		return nil, err
	}

	return pgx.CollectExactlyOneRow(rows, scanVerification)
}
