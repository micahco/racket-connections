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

type LatestPost struct {
	PostID     int
	SkillLevel int
	CreatedAt  time.Time
	UserID     int
	Username   string
	SportName  string
}

func scanLatestPosts(row pgx.CollectableRow) (*LatestPost, error) {
	var p LatestPost
	err := row.Scan(
		&p.PostID,
		&p.SkillLevel,
		&p.CreatedAt,
		&p.UserID,
		&p.Username,
		&p.SportName,
	)
	return &p, err
}

type PostData struct {
	PostID     int
	SkillLevel int
	CreatedAt  time.Time
	UserID     int
	Username   string
}

func (m *PostModel) Latest() (map[string][]*PostData, error) {
	sql := "SELECT * FROM latest_posts;"

	rows, err := m.Pool.Query(context.Background(), sql)
	if err != nil {
		return nil, err
	}

	posts, err := pgx.CollectRows(rows, scanLatestPosts)
	if err != nil {
		return nil, err
	}

	// Key: sport_name
	pm := make(map[string][]*PostData)

	for _, p := range posts {
		// Make array at index if uninitialized
		_, ok := pm[p.SportName]
		if !ok {
			pm[p.SportName] = make([]*PostData, 0)
		}

		pd := &PostData{
			PostID:     p.PostID,
			SkillLevel: p.SkillLevel,
			CreatedAt:  p.CreatedAt,
			UserID:     p.UserID,
			Username:   p.Username,
		}

		pm[p.SportName] = append(pm[p.SportName], pd)
	}

	return pm, nil
}
