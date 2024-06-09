package main

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
)

type session struct {
	id     string
	userID int
	expiry time.Time
}

func (s session) isExpired() bool {
	return s.expiry.Before(time.Now())
}

func scanSession(row pgx.CollectableRow) (*session, error) {
	var s session
	err := row.Scan(
		&s.id,
		&s.userID,
		&s.expiry)
	return &s, err
}

func (db *pgStore) createSession(uuid string, userID int, expiry time.Time) (*session, error) {
	sql := `INSERT INTO sessions
		(id, user_id, expiry)
		VALUES($1, $2, $3) RETURNING *;`
	rows, err := db.Query(context.Background(), sql,
		uuid, userID, expiry)
	if err != nil {
		return nil, err
	}

	return pgx.CollectExactlyOneRow(rows, scanSession)
}

func (db *pgStore) getSession(uuid string) (*session, error) {
	sql := "SELECT * FROM sessions WHERE id = $1;"

	rows, err := db.Query(context.Background(), sql, uuid)
	if err != nil {
		return nil, err
	}

	session, err := pgx.CollectOneRow(rows, scanSession)
	if errors.Is(err, pgx.ErrNoRows) {
		err = fmt.Errorf("no session: %s", uuid)
		return nil, err
	}

	return session, err
}
