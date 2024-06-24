package models

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ContactModel struct {
	pool *pgxpool.Pool
}

func NewContactModel(pool *pgxpool.Pool) *ContactModel {
	return &ContactModel{pool}
}

type UserContact struct {
	ID              int
	UserID          int
	ContactMethodID int
	Value           string
}

func scanUserContact(row pgx.CollectableRow) (*UserContact, error) {
	var c UserContact
	err := row.Scan(
		&c.ID,
		&c.UserID,
		&c.ContactMethodID,
		&c.Value)

	return &c, err
}

func (m *ContactModel) Insert(userID, methodID int, value string) error {
	sql := `INSERT INTO user_contacts 
		(user_id, contact_method_id, value)
		VALUES($1, $2, $3);`

	_, err := m.pool.Exec(context.Background(), sql,
		userID, methodID, value)

	return err
}

func (m *ContactModel) Get(userID int) (*UserContact, error) {
	sql := "SELECT * FROM user_contacts WHERE user_id = $1;"

	rows, err := m.pool.Query(context.Background(), sql, userID)
	if err != nil {
		return nil, err
	}

	c, err := pgx.CollectOneRow(rows, scanUserContact)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNoRecord
		} else {
			return nil, err
		}
	}

	return c, nil
}

func (m *ContactModel) Exists(userID int) (bool, error) {
	var exists bool

	sql := "SELECT EXISTS(SELECT true FROM user_contacts WHERE user_id = $1);"

	err := m.pool.QueryRow(context.Background(), sql, userID).Scan(&exists)

	return exists, err
}

type ContactMethod struct {
	ID   int
	Name string
}

func scanContactMethod(row pgx.CollectableRow) (*ContactMethod, error) {
	var s ContactMethod
	err := row.Scan(&s.ID, &s.Name)

	return &s, err
}

func (m *ContactModel) AllMethods() ([]*ContactMethod, error) {
	sql := "SELECT * FROM contact_methods;"

	rows, err := m.pool.Query(context.Background(), sql)
	if err != nil {
		return nil, err
	}

	return pgx.CollectRows(rows, scanContactMethod)
}
