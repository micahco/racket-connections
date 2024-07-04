package models

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TimeslotModel struct {
	pool *pgxpool.Pool
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

type Timeslot struct {
	Day  *DayOfWeek
	Time *TimeOfDay
}

// Returns a map with key of DayOfWeek ID and values list of TimeOfDay
func (m *TimeslotModel) User(userID int) ([]*Timeslot, error) {
	sql := `SELECT
			d.id_,
			d.name_,
			d.abbrev_,
			t.id_,
			t.name_,
			t.abbrev_
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

	var timeslots []*Timeslot
	for rows.Next() {
		var d DayOfWeek
		var t TimeOfDay

		err := rows.Scan(
			&d.ID,
			&d.Name,
			&d.Abbrev,
			&t.ID,
			&t.Name,
			&t.Abbrev)

		if err != nil {
			return nil, err
		}

		timeslots = append(timeslots, &Timeslot{
			Day:  &d,
			Time: &t,
		})
	}

	return timeslots, nil
}

func (m *TimeslotModel) DeleteUser(userID int) error {
	sql := "DELETE FROM timeslot_ WHERE user_id_ = $1;"

	_, err := m.pool.Exec(context.Background(), sql, userID)

	return err
}
