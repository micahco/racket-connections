package models

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostModel struct {
	pool *pgxpool.Pool
}

type Post struct {
	ID           int
	Comment      string
	CreatedAt    time.Time
	UserID       int
	SportID      int
	SkillLevelID int
}

func (m *PostModel) Insert(userID, sportID, skillLevelID int, comment string) (int, error) {
	var id int

	sql := `INSERT INTO post_
		(user_id_, sport_id_, skill_level_id_, comment_)
		VALUES($1, $2, $3, $4) RETURNING id_;`

	err := m.pool.QueryRow(context.Background(), sql,
		userID, sportID, skillLevelID, comment).Scan(&id)

	return id, err
}

func (m *PostModel) GetID(userID, sportID int) (int, error) {
	var id int

	sql := `SELECT id_ FROM post_ WHERE
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
	Sport      string
	UserName   string
	SkillLevel string
}

func scanPostCard(row pgx.CollectableRow) (*PostCard, error) {
	var p PostCard
	err := row.Scan(
		&p.ID,
		&p.CreatedAt,
		&p.Sport,
		&p.UserName,
		&p.SkillLevel)
	return &p, err
}

func (m *PostModel) Fetch(sports []string, timeslots []Timeslot, limit int, offset int) (int64, []*PostCard, error) {
	sql := `SELECT DISTINCT
				post_.id_,
				post_.created_at_,
				sport_.name_,
				user_.name_,
				skill_level_.name_
			FROM post_
			INNER JOIN sport_
				ON sport_.id_ = post_.sport_id_
			INNER JOIN user_
				ON user_.id_ = post_.user_id_
			INNER JOIN skill_level_
				ON skill_level_.id_ = post_.skill_level_id_`

	if len(timeslots) != 0 {
		sql += `
			INNER JOIN timeslot_
				ON timeslot_.user_id_ = user_.id_
			INNER JOIN day_of_week_
				ON day_of_week_.id_ = timeslot_.day_id_
			INNER JOIN time_of_day_
				ON time_of_day_.id_ = timeslot_.time_id_`
	}

	var args []any
	if len(sports) != 0 || len(timeslots) != 0 {
		sql += "\nWHERE\n"

		if len(sports) != 0 {
			sql += "sport_.name_ = ANY ($1)"
			args = append(args, sports)
			if len(timeslots) != 0 {
				sql += "\nAND "
			}
		}

		if len(timeslots) != 0 {
			for i, t := range timeslots {
				if i != 0 {
					sql += " OR\n"
				}

				idx := len(args)
				sql += fmt.Sprintf(`day_of_week_.abbrev_ = $%d AND time_of_day_.abbrev_ = $%d`, idx+1, idx+2)
				args = append(args, t.Day.Abbrev, t.Time.Abbrev)
			}
		}
	}

	// First query to count number of total rows with filters
	ct, err := m.pool.Exec(context.Background(), sql+";", args...)
	if err != nil {
		return -1, nil, err
	}

	sql += "\nORDER BY post_.id_ DESC\n"

	idx := len(args)
	sql += fmt.Sprintf("LIMIT $%d OFFSET $%d;", idx+1, idx+2)
	args = append(args, limit, offset)

	rows, err := m.pool.Query(context.Background(), sql, args...)
	if err != nil {
		return -1, nil, err
	}

	p, err := pgx.CollectRows(rows, scanPostCard)

	return ct.RowsAffected(), p, err
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

	p, err := pgx.CollectOneRow(rows, scanPostDetails)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNoRecord
		}

		return nil, err
	}

	return p, nil
}
