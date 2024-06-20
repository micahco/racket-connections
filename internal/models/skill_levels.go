package models

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type SkillModel struct {
	pool *pgxpool.Pool
}

func NewSkillModel(pool *pgxpool.Pool) *SkillModel {
	return &SkillModel{pool}
}

type SkillLevel struct {
	Value int
	Desc  string
}

func scanSkill(row pgx.CollectableRow) (*SkillLevel, error) {
	var s SkillLevel
	err := row.Scan(&s.Value, &s.Desc)

	return &s, err
}

func (m *SkillModel) All() ([]*SkillLevel, error) {
	sql := "SELECT * FROM skill_levels;"

	rows, err := m.pool.Query(context.Background(), sql)
	if err != nil {
		return nil, err
	}

	return pgx.CollectRows(rows, scanSkill)
}
