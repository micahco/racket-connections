package main

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
)

func scanSession(row pgx.CollectableRow) (*Session, error) {
	var s Session
	err := row.Scan(
		&s.sessionID,
		&s.userID,
		&s.expiry)
	return &s, err
}

func (s *PostgresStore) CreateSession(sessionID string, userID int, expiry time.Time) (*Session, error) {
	sql := `INSERT INTO sessions
		(session_id, user_id, expiry)
		VALUES($1, $2, $3) RETURNING *;`
	rows, err := s.p.Query(context.Background(), sql,
		sessionID, userID, expiry)
	if err != nil {
		return nil, err
	}

	return pgx.CollectExactlyOneRow(rows, scanSession)
}

func (s *PostgresStore) GetSession(id string) (*Session, error) {
	sql := "SELECT user_id FROM sessions WHERE session_id = $1;"

	rows, err := s.p.Query(context.Background(), sql, id)
	if err != nil {
		return nil, err
	}

	session, err := pgx.CollectOneRow(rows, scanSession)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, err
	}

	return session, err
}
