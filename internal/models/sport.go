package models

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type SportModel struct {
	pool *pgxpool.Pool
}

func NewSportModel(pool *pgxpool.Pool) *SportModel {
	return &SportModel{pool}
}

type Sport struct {
	ID   int
	Name string
}

func scanSport(row pgx.CollectableRow) (*Sport, error) {
	var s Sport
	err := row.Scan(&s.ID, &s.Name)

	return &s, err
}

func (m *SportModel) All() ([]*Sport, error) {
	sql := "SELECT * FROM sport_;"

	rows, err := m.pool.Query(context.Background(), sql)
	if err != nil {
		return nil, err
	}

	return pgx.CollectRows(rows, scanSport)
}

func (m *SportModel) GetID(name string) (int, error) {
	var id int

	sql := "SELECT id_ FROM sport_ WHERE name_ = $1;"

	err := m.pool.QueryRow(context.Background(), sql, name).Scan(&id)

	return id, err
}
