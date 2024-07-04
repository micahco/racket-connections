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

type UserModel struct {
	pool *pgxpool.Pool
}

type User struct {
	ID           int
	Name         string
	Email        string
	PasswordHash []byte
	CreatedAt    time.Time
}

func scanUser(row pgx.CollectableRow) (*User, error) {
	var u User
	err := row.Scan(
		&u.ID,
		&u.Name,
		&u.Email,
		&u.PasswordHash,
		&u.CreatedAt)

	return &u, err
}

func (m *UserModel) Insert(name, email, password string) (int, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return 0, err
	}

	var id int

	sql := `INSERT INTO user_ 
		(name_, email_, password_hash_)
		VALUES($1, $2, $3) RETURNING id_;`

	err = m.pool.QueryRow(context.Background(), sql,
		name, email, hash).Scan(&id)

	if pgErrCode(err) == pgerrcode.UniqueViolation {
		return 0, ErrDuplicateEmail
	}

	return id, err
}

func (m *UserModel) Authenticate(email, password string) (int, error) {
	sql := "SELECT * FROM user_ WHERE email_ = $1;"

	rows, err := m.pool.Query(context.Background(), sql, email)
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

	sql := "SELECT EXISTS(SELECT true FROM user_ WHERE id_ = $1);"

	err := m.pool.QueryRow(context.Background(), sql, id).Scan(&exists)

	return exists, err
}

func (m *UserModel) ExistsEmail(email string) (bool, error) {
	var exists bool

	sql := "SELECT EXISTS(SELECT true FROM user_ WHERE email_ = $1);"

	err := m.pool.QueryRow(context.Background(), sql, email).Scan(&exists)

	return exists, err
}

func (m *UserModel) UpdatePassword(email, password string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	sql := "UPDATE user_ SET password_hash_ = $1 WHERE email_ = $2;"

	_, err = m.pool.Exec(context.Background(), sql, hash, email)

	return err
}

type UserProfile struct {
	Name  string
	Email string
}

func scanUserProfile(row pgx.CollectableRow) (*UserProfile, error) {
	var u UserProfile
	err := row.Scan(&u.Name, &u.Email)

	return &u, err
}

func (m *UserModel) GetProfile(id int) (*UserProfile, error) {
	sql := "SELECT name_, email_ FROM user_ WHERE id_ = $1;"

	rows, err := m.pool.Query(context.Background(), sql, id)
	if err != nil {
		return nil, err
	}

	return pgx.CollectOneRow(rows, scanUserProfile)
}

func (m *UserModel) Delete(id int) error {
	sql := "DELETE FROM user_ WHERE id_ = $1;"

	_, err := m.pool.Exec(context.Background(), sql, id)

	return err
}
