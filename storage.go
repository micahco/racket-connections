package main

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type store interface {
	// Users
	createUser(req signupRequest, hash []byte) (*user, error)
	getUsers() ([]*user, error)
	getUserByID(id int) (*user, error)
	getUserByEmail(email string) (*user, error)
	updateUser(user *user) error
	deleteUser(id int) error

	// Sessions
	createSession(uuid string, userID int, expiry time.Time) (*session, error)
	getSession(uuid string) (*session, error)
}

type pgStore struct {
	*pgxpool.Pool
}

func newPgStore(dsn string) (*pgStore, error) {
	config, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, err
	}

	dbpool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		return nil, err
	}

	if err := dbpool.Ping(context.Background()); err != nil {
		return nil, err
	}

	return &pgStore{dbpool}, nil
}

func pgErrCode(err error) string {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code
	}

	return ""
}

func (db *pgStore) init() error {
	// reset
	db.Exec(context.Background(), "DROP SCHEMA public CASCADE;")
	db.Exec(context.Background(), "CREATE SCHEMA public;")
	db.Exec(context.Background(), "CREATE EXTENSION citext;")

	if err := db.createUsersTable(); err != nil {
		return err
	}
	if err := db.createSessionsTable(); err != nil {
		return err
	}
	if err := db.createSportsTable(); err != nil {
		return err
	}
	if err := db.createPostsTable(); err != nil {
		return err
	}

	return nil
}

func (db *pgStore) createUsersTable() error {
	sql := `CREATE TABLE IF NOT EXISTS users (
		id BIGSERIAL PRIMARY KEY,
		created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
		name TEXT NOT NULL,
		email CITEXT UNIQUE NOT NULL,
		password_hash VARCHAR(100) NOT NULL
	);`
	_, err := db.Exec(context.Background(), sql)
	return err
}

func (db *pgStore) createSessionsTable() error {
	sql := `CREATE TABLE IF NOT EXISTS sessions (
		id UUID NOT NULL PRIMARY KEY,
		user_id INT REFERENCES users(id) NOT NULL,
		expiry TIMESTAMPTZ NOT NULL
	);`
	_, err := db.Exec(context.Background(), sql)
	return err
}

func (db *pgStore) createSportsTable() error {
	sql := `CREATE TABLE IF NOT EXISTS sports (
		id BIGSERIAL PRIMARY KEY,
		sport_name TEXT UNIQUE NOT NULL
	);`
	_, err := db.Exec(context.Background(), sql)
	return err
}

func (db *pgStore) createPostsTable() error {
	sql := `CREATE TABLE IF NOT EXISTS posts (
		id BIGSERIAL PRIMARY KEY,
		created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
		user_id INT REFERENCES users(id) NOT NULL,
		sport_id INT REFERENCES sports(id) NOT NULL
	);`
	_, err := db.Exec(context.Background(), sql)
	return err
}
