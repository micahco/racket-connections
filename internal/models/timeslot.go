package models

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TimeslotModel struct {
	pool *pgxpool.Pool
}

func NewTimeslotModel(pool *pgxpool.Pool) *TimeslotModel {
	return &TimeslotModel{pool}
}

type DayOfWeek struct {
	ID     int
	Name   string
	Abbrev string
}

func scanDayOfWeek(row pgx.CollectableRow) (*DayOfWeek, error) {
	var d DayOfWeek
	err := row.Scan(&d.ID, &d.Name, &d.Abbrev)

	return &d, err
}

func (m *TimeslotModel) Days() ([]*DayOfWeek, error) {
	sql := `SELECT * FROM day_of_week_`

	rows, err := m.pool.Query(context.Background(), sql)
	if err != nil {
		return nil, err
	}

	return pgx.CollectRows(rows, scanDayOfWeek)
}

type TimeOfDay struct {
	ID     int
	Name   string
	Abbrev string
}

func scanTimeOfDay(row pgx.CollectableRow) (*TimeOfDay, error) {
	var t TimeOfDay
	err := row.Scan(&t.ID, &t.Name, &t.Abbrev)

	return &t, err
}

func (m *TimeslotModel) Times() ([]*TimeOfDay, error) {
	sql := `SELECT * FROM time_of_day_`

	rows, err := m.pool.Query(context.Background(), sql)
	if err != nil {
		return nil, err
	}

	return pgx.CollectRows(rows, scanTimeOfDay)
}

func (m *TimeslotModel) Insert(userID, dayID, timeID int) error {
	sql := `INSERT INTO timeslot_ 
		(user_id_, day_id_, time_id_)
		VALUES($1, $2, $3);`

	_, err := m.pool.Exec(context.Background(), sql,
		userID, dayID, timeID)

	return err
}

type TimeslotUser struct {
	Day  string
	Time string
}

func scanTimeslotUser(row pgx.CollectableRow) (*TimeslotUser, error) {
	var d TimeslotUser
	err := row.Scan(&d.Day, &d.Time)

	return &d, err
}

func (m *TimeslotModel) User(userID int) ([]*TimeslotUser, error) {
	sql := `SELECT
			d.name_,
			t.name_
		FROM timeslot_ s
		INNER JOIN day_of_week_ d
			ON d.id_ = s.day_id_
		INNER JOIN time_of_day_ t
			ON t.id_ = s.time_id_
		WHERE s.user_id_ = $1;`

	rows, err := m.pool.Query(context.Background(), sql, userID)
	if err != nil {
		return nil, err
	}

	t, err := pgx.CollectRows(rows, scanTimeslotUser)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNoRecord
		} else {
			return nil, err
		}
	}

	return t, nil
}

type TimeslotAbbrev struct {
	Day  string
	Time string
}
