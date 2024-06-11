package main

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

func newPool(dsn string) (*pgxpool.Pool, error) {
	config, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, err
	}

	pool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		return nil, err
	}

	if err := pool.Ping(context.Background()); err != nil {
		return nil, err
	}

	return pool, nil
}

func createTables(pool *pgxpool.Pool) error {
	// reset
	pool.Exec(context.Background(), "DROP SCHEMA public CASCADE;")
	pool.Exec(context.Background(), "CREATE SCHEMA public;")
	pool.Exec(context.Background(), "CREATE EXTENSION citext;")

	if err := createSessionsTable(pool); err != nil {
		return fmt.Errorf("create sessions table: %w", err)
	}
	if err := createUsersTable(pool); err != nil {
		return fmt.Errorf("create users table: %w", err)
	}
	if err := createVerificationsTable(pool); err != nil {
		return fmt.Errorf("create verifications table: %w", err)
	}
	if err := createSportsTable(pool); err != nil {
		return fmt.Errorf("create sports table: %w", err)
	}
	if err := createPostsTable(pool); err != nil {
		return fmt.Errorf("create posts table: %w", err)
	}

	return nil
}

func createSessionsTable(pool *pgxpool.Pool) error {
	sql := `CREATE TABLE sessions (
		token TEXT PRIMARY KEY,
		data BYTEA NOT NULL,
		expiry TIMESTAMPTZ NOT NULL
	);
	CREATE INDEX sessions_expiry_idx ON sessions (expiry);`
	_, err := pool.Exec(context.Background(), sql)
	return err
}

func createUsersTable(pool *pgxpool.Pool) error {
	sql := `CREATE TABLE IF NOT EXISTS users (
		id BIGSERIAL PRIMARY KEY,
		created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
		is_verified BOOLEAN DEFAULT FALSE,
		name TEXT NOT NULL,
		email CITEXT UNIQUE NOT NULL,
		password_hash CHAR(60) NOT NULL
	);`
	_, err := pool.Exec(context.Background(), sql)
	return err
}

func createVerificationsTable(pool *pgxpool.Pool) error {
	sql := `CREATE TABLE IF NOT EXISTS verifications (
		email CITEXT NOT NULL PRIMARY KEY,
		token VARCHAR(100) NOT NULL,
		expiry TIMESTAMPTZ NOT NULL,
		FOREIGN KEY (email) REFERENCES users(email) ON DELETE CASCADE
	);`
	_, err := pool.Exec(context.Background(), sql)
	return err
}

func createSportsTable(pool *pgxpool.Pool) error {
	sql := `CREATE TABLE IF NOT EXISTS sports (
		id BIGSERIAL PRIMARY KEY,
		sport_name TEXT UNIQUE NOT NULL
	);`
	_, err := pool.Exec(context.Background(), sql)
	return err
}

func createPostsTable(pool *pgxpool.Pool) error {
	sql := `CREATE TABLE IF NOT EXISTS posts (
		id BIGSERIAL PRIMARY KEY,
		created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
		user_id INT NOT NULL,
		sport_id INT NOT NULL,
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
		FOREIGN KEY (sport_id) REFERENCES sports(id) ON DELETE CASCADE
	);`
	_, err := pool.Exec(context.Background(), sql)
	return err
}
