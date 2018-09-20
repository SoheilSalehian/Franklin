package main

import (
	"database/sql"
	"errors"

	"github.com/prometheus/common/log"
)

type User struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Password string `json:"password"`
}

func (u *User) createUser(db *sql.DB) error {
	statement := `INSERT INTO users(name) VALUES('?')`
	_, err := db.Exec(statement, u.Name)
	return err
}

func (u *User) updateUser(db *sql.DB) error {
	return errors.New("TBD")
}

func (u *User) getUser(db *sql.DB) error {
	statement := `SELECT name FROM users WHERE id=$1`
	return db.QueryRow(statement, u.ID).Scan(&u.Name)
}

func (u *User) deleteUser(db *sql.DB) error {
	return errors.New("TBD")
}

type Order struct {
	ID     int    `json:"id",omitempty`
	User   string `json:"user"`
	UserID int    `json:"user_id"`
	Items  `json:"items"`
}

type Items []Item

type Item struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

func (o *Order) createOrder(db *sql.DB) error {
	// log.Info(o)
	statement := `INSERT INTO orders(user_id) VALUES('?')`
	result, err := db.Exec(statement, o.UserID)
	if err != nil {
		log.Error("inserting to orders failed.")
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}

	o.ID = int(id)

	statement = `INSERT INTO order_items(order_id, item_id) VALUES($1, $2)`
	for _, item := range o.Items {
		_, err = db.Exec(statement, o.ID, item.ID)
		if err != nil {
			log.Error("inserting to orders_items failed.")
			return err
		}
	}

	return nil
}

func (o *Order) getOrder(db *sql.DB, userID string) error {

	statement := `SELECT users.name, order_items.item_id, items.name FROM orders 
  INNER JOIN order_items ON order_items.order_id=orders.id
  INNER JOIN items ON order_items.item_id=items.id
  INNER JOIN users ON orders.user_id=users.id
  WHERE orders.id=$1 AND users.id=$2
  `

	rows, err := db.Query(statement, o.ID, userID)
	if err != nil {
		log.Error(err)
		return err
	}
	defer rows.Close()

	i := Item{}

	if rows.Next() {
		err = rows.Scan(&o.User, &i.ID, &i.Name)
		if err != nil {
			log.Error(err)
			return err
		}
		o.Items = append(o.Items, i)
		for rows.Next() {
			err = rows.Scan(&o.User, &i.ID, &i.Name)
			if err != nil {
				log.Error(err)
				return err
			}
			o.Items = append(o.Items, i)
		}
	} else {
		e := errors.New("No DB results found")
		log.Error(e)
		return e
	}

	return nil
}

func getOrders(db *sql.DB, userID string, count, start int) ([]int, error) {

	statement := `SELECT orders.id FROM orders 
  INNER JOIN users ON orders.user_id=users.id
  WHERE users.id=$1 ORDER BY orders.id DESC;
  `
	rows, err := db.Query(statement, userID)
	if err != nil {
		log.Error(err)
		if err == sql.ErrNoRows {
			e := errors.New("No DB results found")
			log.Error(e)
			return nil, err
		}
	}
	defer rows.Close()

	var orders []int
	var id int
	// FIXME: Isn't there a cleaner way?
	if rows.Next() {
		err = rows.Scan(&id)
		if err != nil {
			log.Error(err)
			return nil, err
		}

		orders = append(orders, id)

		for rows.Next() {
			err = rows.Scan(&id)
			if err != nil {
				log.Error(err)
				return nil, err
			}

			orders = append(orders, id)
		}
	} else {
		e := errors.New("No DB results found")
		log.Error(e)
		return nil, e
	}

	return orders, nil
}

func (o *Order) updateOrder(db *sql.DB) error {
	return errors.New("TBD")
}

func (o *Order) deleteOrder(db *sql.DB) error {
	return errors.New("TBD")
}
