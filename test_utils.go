package main

import (
	"fmt"

	"github.com/prometheus/common/log"
	"golang.org/x/crypto/bcrypt"
)

func clearUsersTable() {
	_, err := a.DB.Exec("DELETE FROM users")
	if err != nil {
		log.Error(err)
	}

	_, err = a.DB.Exec("UPDATE SQLITE_SEQUENCE SET SEQ=0 WHERE NAME='users';")
	if err != nil {
		log.Error(err)
	}
}

func clearOrdersTable() {
	_, err := a.DB.Exec("DELETE FROM orders")
	if err != nil {
		log.Error(err)
	}

	_, err = a.DB.Exec("UPDATE SQLITE_SEQUENCE SET SEQ=0 WHERE NAME='orders';")
	if err != nil {
		log.Error(err)
	}
}

func clearOrderItemsTable() {
	_, err := a.DB.Exec("DELETE FROM order_items")
	if err != nil {
		log.Error(err)
	}

	_, err = a.DB.Exec("UPDATE SQLITE_SEQUENCE SET SEQ=0 WHERE NAME='order_items';")
	if err != nil {
		log.Error(err)
	}
}

func clearItemsTable() {
	_, err := a.DB.Exec("DELETE FROM items")
	if err != nil {
		log.Error(err)
	}

	_, err = a.DB.Exec("UPDATE SQLITE_SEQUENCE SET SEQ=0 WHERE NAME='items';")
	if err != nil {
		log.Error(err)
	}
}

func setAuthentication() {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("correct-password"), 8)
	if err != nil {
		log.Error(err)
	}

	statement := fmt.Sprintf(`INSERT INTO users(name,password) VALUES('%s', '%s')`, "Test User", hashedPassword)
	_, err = a.DB.Exec(statement)
	if err != nil {
		log.Error(err)
	}

}
