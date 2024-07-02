package models

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	expiration = time.Hour * 24
)

type VerificationModel struct {
	pool *pgxpool.Pool
}

type Verification struct {
	Token     string
	Email     string
	Expiry    time.Time
	CreatedAt time.Time
}

func (v *Verification) IsExpired() bool {
	return time.Now().After(v.Expiry)
}

func scanVerification(row pgx.CollectableRow) (*Verification, error) {
	var v Verification
	err := row.Scan(
		&v.Token,
		&v.Email,
		&v.Expiry,
		&v.CreatedAt)

	return &v, err
}

func (m *VerificationModel) Insert(token, email string) error {
	expiry := time.Now().Add(expiration)

	sql := `INSERT INTO verification_
		(token_, email_, expiry_)
		VALUES($1, $2, $3);`

	_, err := m.pool.Exec(context.Background(), sql, token, email, expiry)

	return err
}

func (m *VerificationModel) Get(email string) (*Verification, error) {
	sql := "SELECT * FROM verification_ WHERE email_ = $1;"

	rows, err := m.pool.Query(context.Background(), sql, email)
	if err != nil {
		return nil, err
	}

	v, err := pgx.CollectOneRow(rows, scanVerification)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNoRecord
	}

	return v, err
}

func (m *VerificationModel) Verify(token, email string) error {
	sql := `SELECT * FROM verification_ 
		WHERE token_ = $1 AND email_ = $2;`

	rows, err := m.pool.Query(context.Background(), sql, token, email)
	if err != nil {
		return err
	}

	v, err := pgx.CollectOneRow(rows, scanVerification)
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrNoRecord
	}

	if v.IsExpired() {
		return ErrExpiredVerification
	}

	return nil
}

func (m *VerificationModel) Purge(email string) error {
	sql := "DELETE FROM verification_ WHERE email_ = $1;"

	_, err := m.pool.Exec(context.Background(), sql, email)

	return err
}
