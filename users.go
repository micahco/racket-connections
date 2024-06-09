package main

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
)

type user struct {
	id           int
	createdAt    time.Time
	name         string
	email        string
	passwordHash string
}

func scanUser(row pgx.CollectableRow) (*user, error) {
	var u user
	err := row.Scan(
		&u.id,
		&u.createdAt,
		&u.name,
		&u.email,
		&u.passwordHash)

	return &u, err
}

func (db *pgStore) createUser(req signupRequest, hash []byte) (*user, error) {
	sql := `INSERT INTO users
		(name, email, password_hash)
		VALUES($1, $2, $3) RETURNING *;`
	rows, err := db.Query(context.Background(), sql,
		req.name, req.email, hash)
	if err != nil {
		return nil, err
	}

	user, err := pgx.CollectExactlyOneRow(rows, scanUser)
	if pgErrCode(err) == pgerrcode.UniqueViolation {
		err = fmt.Errorf("user already exists: %w", err)
		return nil, err
	}

	return user, err
}

func (db *pgStore) getUsers() ([]*user, error) {
	sql := "SELECT * FROM users"
	rows, err := db.Query(context.Background(), sql)
	if err != nil {
		return nil, err
	}

	return pgx.CollectRows(rows, scanUser)
}

func (db *pgStore) getUserByID(id int) (*user, error) {
	sql := "SELECT * FROM users WHERE id = $1;"
	rows, err := db.Query(context.Background(), sql, id)
	if err != nil {
		return nil, err
	}

	user, err := pgx.CollectOneRow(rows, scanUser)
	if errors.Is(err, pgx.ErrNoRows) {
		err = fmt.Errorf("no user with id: %d", id)
		return nil, err
	}

	return user, err
}

func (db *pgStore) getUserByEmail(email string) (*user, error) {
	sql := "SELECT * FROM users WHERE email = $1;"
	rows, err := db.Query(context.Background(), sql, email)
	if err != nil {
		return nil, err
	}

	user, err := pgx.CollectOneRow(rows, scanUser)
	if errors.Is(err, pgx.ErrNoRows) {
		err = fmt.Errorf("no user with email: %s", email)
		return nil, err
	}

	return user, err
}

func (db *pgStore) updateUser(user *user) error {
	return nil
}

func (db *pgStore) deleteUser(id int) error {
	sql := "DELETE FROM users WHERE id = $1;"
	commandTag, err := db.Exec(context.Background(), sql, id)
	if err != nil {
		return err
	}

	if commandTag.RowsAffected() != 1 {
		err = fmt.Errorf("no user with id: %d", id)
		return err
	}

	return err
}
