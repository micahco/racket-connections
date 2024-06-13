package models

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Post struct {
	ID      int
	UserID  int
	SportID int
	Created time.Time
}

type PostModel struct {
	Pool *pgxpool.Pool
}

func scanPost(row pgx.CollectableRow) (*Post, error) {
	var p Post
	err := row.Scan(
		&p.ID,
		&p.UserID,
		&p.SportID,
		&p.Created)
	return &p, err
}

func (m *PostModel) Insert(userID, sportID int) (*Post, error) {
	sql := `INSERT INTO posts
		(user_id, sport_id)
		VALUES($1, $2) RETURNING *;`

	rows, err := m.Pool.Query(context.Background(), sql, userID, sportID)
	if err != nil {
		return nil, err
	}

	return pgx.CollectOneRow(rows, scanPost)
}

func (m *PostModel) Get(id int) (*Post, error) {
	sql := "SELECT * FROM posts WHERE id = $1;"

	rows, err := m.Pool.Query(context.Background(), sql, id)
	if err != nil {
		return nil, err
	}

	p, err := pgx.CollectOneRow(rows, scanPost)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNoRecord
	}

	return p, err
}

func (m *PostModel) Latest(count int) ([]*Post, error) {
	sql := "SELECT * FROM posts ORDER BY id DESC LIMIT $1"

	rows, err := m.Pool.Query(context.Background(), sql, count)
	if err != nil {
		return nil, err
	}

	return pgx.CollectRows(rows, scanPost)
}
