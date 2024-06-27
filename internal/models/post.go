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

type Post struct {
	ID           int
	Comment      string
	CreatedAt    time.Time
	UserID       int
	SportID      int
	SkillLevelID int
}

func (m *PostModel) Insert(comment string, userID, sportID, skillLevelID int) (int, error) {
	var id int

	sql := `INSERT INTO post_
		(comment_, user_id_, sport_id_, skill_level_id_)
		VALUES($1, $2, $3, $4) RETURNING id;`

	err := m.pool.QueryRow(context.Background(), sql,
		comment, userID, sportID, skillLevelID).Scan(&id)

	return id, err
}

func (m *PostModel) GetID(userID, sportID int) (int, error) {
	var id int

	sql := `SELECT id FROM post_ WHERE
		user_id_ = $1 AND sport_id_ = $2;`

	err := m.pool.QueryRow(context.Background(), sql, userID, sportID).Scan(&id)
	if errors.Is(err, pgx.ErrNoRows) {
		return 0, ErrNoRecord
	}

	return id, err
}

func (m *PostModel) GetUserID(id int) (int, error) {
	var userID int

	sql := "SELECT user_id_ FROM post_ WHERE id_ = $1;"

	err := m.pool.QueryRow(context.Background(), sql, id).Scan(&userID)

	return userID, err
}

func (m *PostModel) Delete(id int) error {
	sql := "DELETE FROM post_ WHERE id_ = $1;"

	_, err := m.pool.Exec(context.Background(), sql, id)

	return err
}

type PostCard struct {
	ID         int
	CreatedAt  time.Time
	UserName   string
	SkillLevel string
}

func (m *PostModel) All(sports []string) (map[int][]PostCard, error) {
	sql := `SELECT
			s.id_,
			p.id_,
			p.created_at_,
			u.name_,
			l.name_
		FROM post_ p	
		INNER JOIN sport_ s
			ON s.id_ = p.sport_id_
		INNER JOIN user_ u
			ON u.id_ = p.user_id_
		INNER JOIN skill_level_ l
			ON l.id_ = p.skill_level_id_`

	var rows pgx.Rows
	var err error
	if len(sports) != 0 {
		sql += "\nWHERE s.name_ = ANY ($1);"
		rows, err = m.pool.Query(context.Background(), sql, sports)
	} else {
		sql += ";"
		rows, err = m.pool.Query(context.Background(), sql)
	}

	if err != nil {
		return nil, err
	}

	postsMap := make(map[int][]PostCard)
	for rows.Next() {
		var sportID int
		var post PostCard

		err := rows.Scan(
			&sportID,
			&post.ID,
			&post.CreatedAt,
			&post.UserName,
			&post.SkillLevel)
		if err != nil {
			return nil, err
		}

		postsMap[sportID] = append(postsMap[sportID], post)
	}

	return postsMap, nil
}

func (m *PostModel) Latest(count int) (map[int][]PostCard, error) {
	sql := `SELECT
			s.id_,
			p.id_,
			p.created_at_,
			u.name_,
			l.name_
		FROM v_post_numbered_ p
		INNER JOIN sport_ s
			ON s.id_ = p.sport_id_
		INNER JOIN user_ u
			ON u.id_ = p.user_id_
		INNER JOIN skill_level_ l
			ON l.id_ = p.skill_level_id_
		WHERE ROW_NUMBER <= $1;`

	rows, err := m.pool.Query(context.Background(), sql, count)
	if err != nil {
		return nil, err
	}

	postsMap := make(map[int][]PostCard)
	for rows.Next() {
		var sportID int
		var post PostCard

		err := rows.Scan(
			&sportID,
			&post.ID,
			&post.CreatedAt,
			&post.UserName,
			&post.SkillLevel)
		if err != nil {
			return nil, err
		}

		postsMap[sportID] = append(postsMap[sportID], post)
	}

	return postsMap, nil
}

type PostDetails struct {
	ID         int
	Comment    string
	CreatedAt  time.Time
	UserID     int
	UserName   string
	Sport      string
	SkillLevel string
}

func scanPostDetails(row pgx.CollectableRow) (*PostDetails, error) {
	var p PostDetails
	err := row.Scan(
		&p.ID,
		&p.Comment,
		&p.CreatedAt,
		&p.UserID,
		&p.UserName,
		&p.Sport,
		&p.SkillLevel)
	return &p, err
}

func (m *PostModel) GetDetails(id int) (*PostDetails, error) {
	sql := `SELECT
			p.id_,
			p.comment_,
			p.created_at_,
			u.id_,
			u.name_,
			s.name_,
			l.name_
		FROM post_ p
		INNER JOIN user_ u
			ON u.id_ = p.user_id_
		INNER JOIN sport_ s
			ON s.id_ = p.sport_id_
		INNER JOIN skill_level_ l
			ON l.id_ = p.skill_level_id_
		WHERE p.id_ = $1;`

	rows, err := m.pool.Query(context.Background(), sql, id)
	if err != nil {
		return nil, err
	}

	return pgx.CollectOneRow(rows, scanPostDetails)
}
