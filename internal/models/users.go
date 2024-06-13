package models

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID           int
	CreatedAt    time.Time
	Name         string
	Email        string
	PasswordHash []byte
}

type UserModel struct {
	Pool *pgxpool.Pool
}

func scanUser(row pgx.CollectableRow) (*User, error) {
	var u User
	err := row.Scan(
		&u.ID,
		&u.CreatedAt,
		&u.Name,
		&u.Email,
		&u.PasswordHash)

	return &u, err
}

func (m *UserModel) Insert(name, email, password string) (int, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return 0, err
	}

	var id int

	sql := `INSERT INTO users 
		(name, email, password_hash)
		VALUES($1, $2, $3) RETURNING id;`

	err = m.Pool.QueryRow(context.Background(), sql, name, email, hash).Scan(&id)
	if pgErrCode(err) == pgerrcode.UniqueViolation {
		return 0, ErrDuplicateEmail
	}

	return id, err
}

func (m *UserModel) Authenticate(email, password string) (int, error) {
	sql := "SELECT * FROM users WHERE email = $1;"

	rows, err := m.Pool.Query(context.Background(), sql, email)
	if err != nil {
		return 0, err
	}

	user, err := pgx.CollectOneRow(rows, scanUser)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, ErrInvalidCredentials
		} else {
			return 0, err
		}
	}

	err = bcrypt.CompareHashAndPassword(user.PasswordHash, []byte(password))
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return -0, ErrInvalidCredentials
		} else {
			return 0, err
		}
	}

	return user.ID, nil
}

func (m *UserModel) Exists(id int) (bool, error) {
	var exists bool

	sql := "SELECT EXISTS(SELECT true FROM users WHERE id = $1);"

	err := m.Pool.QueryRow(context.Background(), sql, id).Scan(&exists)

	return exists, err
}

func (m *UserModel) ExistsEmail(email string) (bool, error) {
	var exists bool

	sql := "SELECT EXISTS(SELECT true FROM users WHERE email = $1);"

	err := m.Pool.QueryRow(context.Background(), sql, email).Scan(&exists)

	return exists, err
}

func (m *UserModel) UpdatePassword(email, password string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	sql := "UPDATE users SET password_hash = $1 WHERE email = $2"

	_, err = m.Pool.Exec(context.Background(), sql, hash, email)

	return err
}
