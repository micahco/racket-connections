package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
)

func (s *PostgresStore) CreateUser(req SignupRequest, hash []byte) (*User, error) {
	sql := `INSERT INTO users
		(email, first_name, last_name, password_hash)
		VALUES($1, $2, $3, $4) RETURNING *;`
	rows, err := s.p.Query(context.Background(), sql,
		req.Email,
		req.FirstName,
		req.LastName,
		hash)
	if err != nil {
		return nil, err
	}

	user, err := pgx.CollectExactlyOneRow(rows, scanUser)
	if pgErrCode(err) == pgerrcode.UniqueViolation {
		err = fmt.Errorf("user already exists: %w", err)
		return nil, NewAPIError(err, http.StatusConflict)
	}

	return user, err
}

func (s *PostgresStore) GetUsers() ([]*User, error) {
	sql := "SELECT * FROM users"
	rows, err := s.p.Query(context.Background(), sql)
	if err != nil {
		return nil, err
	}

	return pgx.CollectRows(rows, scanUser)
}

func (s *PostgresStore) GetUserByID(id int) (*User, error) {
	sql := "SELECT * FROM users WHERE user_id = $1;"
	rows, err := s.p.Query(context.Background(), sql, id)
	if err != nil {
		return nil, err
	}

	user, err := pgx.CollectOneRow(rows, scanUser)
	if errors.Is(err, pgx.ErrNoRows) {
		err = fmt.Errorf("no user with id: %d", id)
		return nil, NewAPIError(err, http.StatusNotFound)
	}

	return user, err
}

func (s *PostgresStore) GetUserByEmail(email string) (*User, error) {
	sql := "SELECT * FROM users WHERE email = $1;"
	rows, err := s.p.Query(context.Background(), sql, email)
	if err != nil {
		return nil, err
	}

	user, err := pgx.CollectOneRow(rows, scanUser)
	if errors.Is(err, pgx.ErrNoRows) {
		err = fmt.Errorf("no user with email: %s", email)
		return nil, NewAPIError(err, http.StatusNotFound)
	}

	return user, err
}

func (s *PostgresStore) UpdateUser(user *User) error {
	return nil
}

func (s *PostgresStore) DeleteUser(id int) error {
	sql := "DELETE FROM users WHERE user_id = $1;"
	commandTag, err := s.p.Exec(context.Background(), sql, id)
	if err != nil {
		return err
	}
	if commandTag.RowsAffected() != 1 {
		err = fmt.Errorf("no user with id: %d", id)
		return NewAPIError(err, http.StatusNotFound)
	}
	return err
}

func scanUser(row pgx.CollectableRow) (*User, error) {
	var user User
	err := row.Scan(
		&user.ID,
		&user.Email,
		&user.FirstName,
		&user.LastName,
		&user.CreatedAt,
		&user.PasswordHash)
	return &user, err
}
