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

func (m *ContactModel) Insert(value string, userID, methodID int) error {
	sql := `INSERT INTO contact_ 
		(value_, user_id_, contact_method_id_)
		VALUES($1, $2, $3);`

	_, err := m.pool.Exec(context.Background(), sql,
		value, userID, methodID)

	return err
}

func (m *ContactModel) Exists(userID int) (bool, error) {
	var exists bool

	sql := "SELECT EXISTS(SELECT true FROM contact_ WHERE user_id_ = $1);"

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

func (m *ContactModel) Methods() ([]*ContactMethod, error) {
	sql := "SELECT * FROM contact_method_;"

	rows, err := m.pool.Query(context.Background(), sql)
	if err != nil {
		return nil, err
	}

	return pgx.CollectRows(rows, scanContactMethod)
}

type UserContact struct {
	ID     int
	Value  string
	Method string
}

func scanUserContact(row pgx.CollectableRow) (*UserContact, error) {
	var c UserContact
	err := row.Scan(&c.ID, &c.Value, &c.Method)

	return &c, err
}

func (m *ContactModel) UserContacts(userID int) ([]*UserContact, error) {
	sql := `SELECT
			c.id_,
			c.value_,
			m.name_
		FROM contact_ c
		INNER JOIN contact_method_ m
			ON c.contact_method_id_ = m.id_
		WHERE c.user_id_ = $1;`

	rows, err := m.pool.Query(context.Background(), sql, userID)
	if err != nil {
		return nil, err
	}

	c, err := pgx.CollectRows(rows, scanUserContact)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNoRecord
		} else {
			return nil, err
		}
	}

	return c, nil
}

func (m *ContactModel) MethodID(name string) (int, error) {
	var id int

	sql := "SELECT id_ FROM contact_method_ WHERE name_ = $1;"

	err := m.pool.QueryRow(context.Background(), sql, name).Scan(&id)

	return id, err
}

func (m *ContactModel) Delete(id int) error {
	sql := "DELETE FROM contact_ WHERE id_ = $1;"

	_, err := m.pool.Exec(context.Background(), sql, id)

	return err
}
