package models

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type SkillLevelModel struct {
	pool *pgxpool.Pool
}

func NewSkillLevelModel(pool *pgxpool.Pool) *SkillLevelModel {
	return &SkillLevelModel{pool}
}

type SkillLevel struct {
	ID   int
	Name string
}

func scanSkillLevel(row pgx.CollectableRow) (*SkillLevel, error) {
	var s SkillLevel
	err := row.Scan(&s.ID, &s.Name)

	return &s, err
}

func (m *SkillLevelModel) All() ([]*SkillLevel, error) {
	sql := "SELECT * FROM skill_level_;"

	rows, err := m.pool.Query(context.Background(), sql)
	if err != nil {
		return nil, err
	}

	return pgx.CollectRows(rows, scanSkillLevel)
}
