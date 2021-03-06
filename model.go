package main

import (
	"database/sql"
	"errors"
	"strconv"

	"github.com/prometheus/common/log"
)

type User struct {
	ID           int    `json:"id"`
	Name         string `json:"name"`
	Password     string `json:"password,omitempty"`
	Zipcode      int    `json:"zipcode,omitempty"`
	ClosestStore Store  `json:"closest_store,omitempty"`
}

type Store struct {
	City          string    `json:"city,omitempty"`
	Coordinates   []float64 `json:"coordinates,omitempty"`
	Country       string    `json:"country,omitempty"`
	Name          string    `json:"name,omitempty"`
	No            int       `json:"no,omitempty"`
	PhoneNumber   string    `json:"phoneNumber,omitempty"`
	StateProvCode string    `json:"stateProvCode,omitempty"`
	StreetAddress string    `json:"streetAddress,omitempty"`
	SundayOpen    bool      `json:"sundayOpen,omitempty"`
	Timezone      string    `json:"timezone,omitempty"`
	Zip           string    `json:"zip,omitempty"`
}

func (u *User) createUser(db *sql.DB) error {
	statement := "INSERT INTO users(name,password,zip,store_lat,store_lon) VALUES(?, ?, ?, ?, ?)"

	// NOTE: For simplicity, we are assuming that only storing the coordinates from the external API call is allowed.
	result, err := db.Exec(statement, u.Name, u.Password, u.Zipcode, u.ClosestStore.Coordinates[0], u.ClosestStore.Coordinates[1])

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}

  log.Info(id)

	u.ID = int(id)
	u.Password = ""
	return err
}

func (u *User) getUser(db *sql.DB) error {
	statement := `SELECT name,store_lat,store_lon FROM users WHERE id=$1`
	var lat, lon float64
	err := db.QueryRow(statement, u.ID).Scan(&u.Name, &lat, &lon)
	u.ClosestStore.Coordinates = append(u.ClosestStore.Coordinates, lat)
	u.ClosestStore.Coordinates = append(u.ClosestStore.Coordinates, lon)
	return err
}

type Order struct {
	ID     int    `json:"id,omitempty"`
	User   string `json:"user"`
	UserID int    `json:"user_id"`
	Items  `json:"items"`
}

type Orders []Order

type Items []Item

type Item struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

func (o *Order) createOrder(db *sql.DB) error {
	statement := `INSERT INTO orders(user_id) VALUES($1)`
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
			log.Error("inserting to order_items failed.")
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

func getOrders(db *sql.DB, userID string, count, start int) (Orders, error) {

	statement := `SELECT orders.id FROM orders 
  INNER JOIN users ON orders.user_id=users.id
  WHERE users.id=$1 ORDER BY orders.id DESC LIMIT $2 OFFSET $3;
  `

	rows, err := db.Query(statement, userID, count, start)
	if err != nil {
		log.Error(err)
		if err == sql.ErrNoRows {
			e := errors.New("No DB results found")
			log.Error(e)
			return nil, err
		}
	}
	defer rows.Close()

	var orders Orders
	var orderIDs []int
	var oID int

	// FIXME: Isn't there a cleaner way? needs serious refactoring
	if rows.Next() {
		err = rows.Scan(&oID)
		if err != nil {
			log.Error(err)
			return nil, err
		}

		orderIDs = append(orderIDs, oID)

		for rows.Next() {
			err = rows.Scan(&oID)
			if err != nil {
				log.Error(err)
				return nil, err
			}

			orderIDs = append(orderIDs, oID)
		}
	} else {
		e := errors.New("No DB results found")
		log.Error(e)
		return nil, e
	}

	for _, oID := range orderIDs {

		statement := `SELECT users.name, order_items.item_id, items.name FROM orders 
  INNER JOIN order_items ON order_items.order_id=orders.id
  INNER JOIN items ON order_items.item_id=items.id
  INNER JOIN users ON orders.user_id=users.id
  WHERE orders.id=$1 AND users.id=$2;
  `

		rows, err := db.Query(statement, oID, userID)
		if err != nil {
			log.Error(err)
			return nil, err
		}
		defer rows.Close()

		o := Order{}
		i := Item{}

		if rows.Next() {
			err = rows.Scan(&o.User, &i.ID, &i.Name)
			if err != nil {
				log.Error(err)
				return nil, err
			}
			o.Items = append(o.Items, i)
			o.ID = oID
			o.UserID, _ = strconv.Atoi(userID)
			for rows.Next() {
				err = rows.Scan(&o.User, &i.ID, &i.Name)
				if err != nil {
					log.Error(err)
					return nil, err
				}
				o.Items = append(o.Items, i)
				o.ID = oID
				o.UserID, _ = strconv.Atoi(userID)
			}
		} else {
			e := errors.New("No DB results found")
			log.Error(e)
			return nil, e
		}

		orders = append(orders, o)
	}

	return orders, nil
}

// FIXME: This needs to be transaction based
func (o *Order) updateOrder(db *sql.DB) error {

	statement := `SELECT item_id FROM order_items WHERE order_id=?`
	rows, err := db.Query(statement, o.ID)
	if err != nil {
		log.Error(err)
		return err
	}
	defer rows.Close()

	var existing []int
	var eid int

	if rows.Next() {
		err = rows.Scan(&eid)
		if err != nil {
			log.Error(err)
			return err
		}

		existing = append(existing, eid)
		for rows.Next() {
			err = rows.Scan(&eid)
			if err != nil {
				log.Error(err)
				return err
			}

			existing = append(existing, eid)
		}

	} else {
		e := errors.New("Order not found.")
		log.Error(e)
		return e
	}

	desired := o.getItemIDs()

	dels := compare(existing, desired)
	adds := compare(desired, existing)

	statement = `INSERT INTO order_items(order_id, item_id) VALUES($1, $2)`
	for _, addID := range adds {
		_, err = db.Exec(statement, o.ID, addID)
		if err != nil {
			log.Error("inserting to order_items failed.")
			return err
		}
	}

	statement = `DELETE FROM order_items WHERE order_id=? AND item_id=?;`
	for _, delID := range dels {
		_, err = db.Exec(statement, o.ID, delID)
		if err != nil {
			log.Error("deleting from order_items failed.")
			return err
		}
	}

	// Verify the results
	statement = `SELECT COUNT(item_id) FROM order_items WHERE order_id=?`
	rows, err = db.Query(statement, o.ID)
	if err != nil {
		log.Error(err)
		return err
	}
	defer rows.Close()

	var count int
	for rows.Next() {
		err = rows.Scan(&count)
		if err != nil {
			log.Error(err)
			return err
		}
	}

	if count != len(desired) {
		e := errors.New("Incorrect updates on order_items.")
		log.Info(e)
		return e
	}

	return nil
}

func (o *Order) getItemIDs() []int {
	var ids []int
	for _, item := range o.Items {
		ids = append(ids, item.ID)
	}
	return ids
}

func (o *Order) deleteOrder(db *sql.DB) error {

	statement := `DELETE FROM orders WHERE user_id =? AND id=?`
	result, err := db.Exec(statement, o.UserID, o.ID)
	if err != nil {
		log.Error("deleting from orders failed: ", err)
		return err
	}

	number, err := result.RowsAffected()
	if err != nil {
		log.Error(err)
		return err
	}

	if int(number) == 0 {
		e := errors.New("Order doesn't exist.")
		log.Error(e)
		return e
	}

	return nil
}
