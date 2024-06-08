package main

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Storage interface {
	// users
	CreateUser(req SignupRequest, hash []byte) (*User, error)
	GetUsers() ([]*User, error)
	GetUserByID(id int) (*User, error)
	GetUserByEmail(email string) (*User, error)
	UpdateUser(user *User) error
	DeleteUser(id int) error

	// sessions
	CreateSession(sessionID string, userID int, expiry time.Time) (*Session, error)
	GetSession(uuid string) (*Session, error)
}

type PostgresStore struct {
	p *pgxpool.Pool
}

func NewPostgresStore(connString string) (*PostgresStore, error) {
	dbpool, err := pgxpool.New(context.Background(), connString)
	if err != nil {
		return nil, err
	}

	if err := dbpool.Ping(context.Background()); err != nil {
		return nil, err
	}

	return &PostgresStore{p: dbpool}, nil
}

func (s *PostgresStore) Init() error {
	// reset
	s.p.Exec(context.Background(), "DROP SCHEMA public CASCADE;")
	s.p.Exec(context.Background(), "CREATE SCHEMA public;")

	if err := s.createUsersTable(); err != nil {
		return err
	}
	if err := s.createSessionsTable(); err != nil {
		return err
	}
	if err := s.createSportsTable(); err != nil {
		return err
	}
	if err := s.createPostsTable(); err != nil {
		return err
	}

	return nil
}

// https://github.com/jackc/pgx/wiki/Error-Handling
func pgErrCode(err error) string {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code
	}

	return ""
}

func (s *PostgresStore) createUsersTable() error {
	sql := `CREATE TABLE IF NOT EXISTS users (
		user_id SERIAL PRIMARY KEY,
		email VARCHAR (200) UNIQUE NOT NULL,
		first_name VARCHAR (200) NOT NULL,
		last_name VARCHAR (200) NOT NULL,
		created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
		password_hash VARCHAR(100) NOT NULL
	);`
	_, err := s.p.Exec(context.Background(), sql)
	return err
}

func (s *PostgresStore) createSessionsTable() error {
	sql := `CREATE TABLE IF NOT EXISTS sessions (
		session_id UUID NOT NULL PRIMARY KEY,
		user_id INT REFERENCES users(user_id) NOT NULL,
		expiry TIMESTAMPTZ NOT NULL
	);`
	_, err := s.p.Exec(context.Background(), sql)
	return err
}

func (s *PostgresStore) createSportsTable() error {
	sql := `CREATE TABLE IF NOT EXISTS sports (
		sport_id SERIAL PRIMARY KEY,
		sport_name VARCHAR(200) UNIQUE NOT NULL
	);`
	_, err := s.p.Exec(context.Background(), sql)
	return err
}

func (s *PostgresStore) createPostsTable() error {
	sql := `CREATE TABLE IF NOT EXISTS posts (
		post_id SERIAL PRIMARY KEY,
		user_id INT REFERENCES users(user_id) NOT NULL,
		sport_id INT REFERENCES sports(sport_id) NOT NULL,
		created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
	);`
	_, err := s.p.Exec(context.Background(), sql)
	return err
}
