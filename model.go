package main

import (
	"database/sql"
	"errors"
)

type user struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

func (u *user) createUser(db *sql.DB) error {
	statement := `INSERT INTO users(name) VALUES('?')`
	_, err := db.Exec(statement, u.Name)
	return err
}

func (u *user) updateUser(db *sql.DB) error {
	return errors.New("TBD")
}

func (u *user) getUser(db *sql.DB) error {
	statement := `SELECT name FROM users WHERE id=$1`
	return db.QueryRow(statement, u.ID).Scan(&u.Name)
}

func (u *user) deleteUser(db *sql.DB) error {
	return errors.New("TBD")
}
