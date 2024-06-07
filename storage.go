package main

import (
	"context"
	"errors"
	"net/http"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Storage interface {
	CreateUser(req CreateUserRequest) (*User, error)
	UpdateUser(user *User) error
	DeleteUser(id int) error
	GetUserByID(id int) (*User, error)
	GetUsers() ([]*User, error)
}

type PostgresStore struct {
	p *pgxpool.Pool
}

func NewPostgresStore() (*PostgresStore, error) {
	connString := "postgres://postgres:postgres@localhost:5432/postgres"
	dbpool, err := pgxpool.New(context.Background(), connString)
	if err != nil {
		return nil, err
	}

	if err := dbpool.Ping(context.Background()); err != nil {
		return nil, err
	}

	return &PostgresStore{
		p: dbpool,
	}, nil
}

func pgErrCode(err error) string {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code
	}
	return ""
}

func scanUser(row pgx.CollectableRow) (*User, error) {
	var user User
	err := row.Scan(
		&user.ID,
		&user.Email,
		&user.FirstName,
		&user.LastName,
		&user.CreatedAt)
	return &user, err
}

func (s *PostgresStore) CreateUser(req CreateUserRequest) (*User, error) {
	sql := `INSERT INTO users 
		(email, first_name, last_name)
		VALUES($1, $2, $3)
		RETURNING *;`
	rows, err := s.p.Query(context.Background(), sql,
		req.Email,
		req.FirstName,
		req.LastName)
	if err != nil {
		return nil, err
	}

	users, err := pgx.CollectExactlyOneRow(rows, scanUser)
	if code := pgErrCode(err); code == "23505" {
		// duplicate key value violates unique constraint
		return nil, NewAPIError(err, http.StatusConflict)
	}

	return users, err
}

func (s *PostgresStore) UpdateUser(user *User) error {
	return nil
}

func (s *PostgresStore) DeleteUser(id int) error {
	return nil
}

func (s *PostgresStore) GetUserByID(id int) (*User, error) {
	return nil, nil
}

func (s *PostgresStore) GetUsers() ([]*User, error) {
	sql := "SELECT * FROM users"
	rows, err := s.p.Query(context.Background(), sql)
	if err != nil {
		return nil, err
	}
	users := []*User{}
	for rows.Next() {
		user := new(User)
		err := rows.Scan(
			&user.ID,
			&user.Email,
			&user.FirstName,
			&user.LastName,
			&user.CreatedAt)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	rows.Close()
	return users, nil
}

func (s *PostgresStore) Init() error {
	// RESET
	s.p.Exec(context.Background(), "DROP SCHEMA public CASCADE;")
	s.p.Exec(context.Background(), "CREATE SCHEMA public;")

	// Create tables
	if err := s.createUsersTable(); err != nil {
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

func (s *PostgresStore) createUsersTable() error {
	sql := `CREATE TABLE IF NOT EXISTS users (
		user_id SERIAL PRIMARY KEY,
		email VARCHAR (200) UNIQUE NOT NULL,
		first_name VARCHAR (200) NOT NULL,
		last_name VARCHAR (200) NOT NULL,
		created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
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
		user_id INT REFERENCES users(user_id),
		sport_id INT REFERENCES sports(sport_id),
		created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
	);`
	_, err := s.p.Exec(context.Background(), sql)
	return err
}
