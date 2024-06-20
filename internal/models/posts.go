package models

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostModel struct {
	pool *pgxpool.Pool
}

func NewPostModel(pool *pgxpool.Pool) *PostModel {
	return &PostModel{pool}
}

type PostDetails struct {
	PostID     int
	SkillLevel int
	CreatedAt  time.Time
	UserID     int
	Username   string
	SportID    int
	SportName  string
}

func scanPostDetails(row pgx.CollectableRow) (*PostDetails, error) {
	var p PostDetails
	err := row.Scan(
		&p.PostID,
		&p.SkillLevel,
		&p.CreatedAt,
		&p.UserID,
		&p.Username,
		&p.SportID,
		&p.SportName)
	return &p, err
}

func (m *PostModel) Insert(userID, sportID, skillLevel int) (int, error) {
	var id int

	sql := `INSERT INTO posts
		(user_id, sport_id, skill_level)
		VALUES($1, $2, $3) RETURNING id;`

	err := m.pool.QueryRow(context.Background(), sql, userID, sportID, skillLevel).Scan(&id)

	return id, err
}

func (m *PostModel) Get(id int) (*PostDetails, error) {
	sql := "SELECT * FROM post_details WHERE id = $1;"

	rows, err := m.pool.Query(context.Background(), sql, id)
	if err != nil {
		return nil, err
	}

	p, err := pgx.CollectOneRow(rows, scanPostDetails)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNoRecord
	}

	return p, err
}

func (m *PostModel) Exists(userID, sportID int) (bool, error) {
	var exists bool

	sql := "SELECT EXISTS(SELECT true FROM posts WHERE user_id = $1 AND sport_id = $2);"

	err := m.pool.QueryRow(context.Background(), sql, userID, sportID).Scan(&exists)

	return exists, err
}

func (m *PostModel) Latest() (map[string][]*PostDetails, error) {
	sql := "SELECT * FROM latest_posts;"

	rows, err := m.pool.Query(context.Background(), sql)
	if err != nil {
		return nil, err
	}

	posts, err := pgx.CollectRows(rows, scanPostDetails)
	if err != nil {
		return nil, err
	}

	// Key: sport_name
	pm := make(map[string][]*PostDetails)

	for _, p := range posts {
		// Make array at index if uninitialized
		_, ok := pm[p.SportName]
		if !ok {
			pm[p.SportName] = make([]*PostDetails, 0)
		}

		pm[p.SportName] = append(pm[p.SportName], p)
	}

	return pm, nil
}
