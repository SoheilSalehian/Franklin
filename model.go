package main

import (
	"database/sql"
	"errors"
	"fmt"
)

type user struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

func (u *user) createUser(db *sql.DB) error {
	return errors.New("TBD")
}

func (u *user) updateUser(db *sql.DB) error {
	return errors.New("TBD")
}

func (u *user) getUser(db *sql.DB) error {
	statement := fmt.Sprintf("SELECT name FROM users WHERE id=%d", u.ID)
	return db.QueryRow(statement).Scan(&u.Name)
}

func (u *user) deleteUser(db *sql.DB) error {
	return errors.New("TBD")
}
