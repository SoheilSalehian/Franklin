package main

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"golang.org/x/crypto/bcrypt"

	"github.com/prometheus/common/log"
	"github.com/stretchr/testify/assert"
)

var a App

func TestMain(m *testing.M) {
	a = App{}
	a.InitDB("franklin-test.db")
	a.InitRouter()

	query :=
		`CREATE TABLE IF NOT EXISTS users (
    id INTEGER AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    password TEXT NOT NULL)`

	if _, err := a.DB.Exec(query); err != nil {
		log.Fatal(err)
	}

	code := m.Run()

	clearUsersTable()
	clearOrdersTable()
	clearOrderItemsTable()
	clearItemsTable()

	os.Exit(code)
}

func TestUserIDDoesNotExist(t *testing.T) {

	clearUsersTable()

	setAuthentication()

	req, _ := http.NewRequest("GET", "/users/15", nil)

	req.SetBasicAuth("Test User", "correct-password")

	response := httptest.NewRecorder()
	a.Router.ServeHTTP(response, req)

	actual := string(response.Body.Bytes())

	assert.JSONEq(t, actual, `{"error":"User not found."}`)

	assert.Equal(t, response.Code, http.StatusNotFound)
}

func TestGetUser(t *testing.T) {
	clearUsersTable()

	setAuthentication()

	_, err := a.DB.Exec("INSERT INTO users(name) VALUES('Test User')")
	if err != nil {
		log.Error(err)
	}

	req, _ := http.NewRequest("GET", "/users/1", nil)

	req.SetBasicAuth("Test User", "correct-password")

	response := httptest.NewRecorder()
	a.Router.ServeHTTP(response, req)

	actual := string(response.Body.Bytes())

	assert.JSONEq(t, actual, `{"id":1,"name":"Test User"}`)

	assert.Equal(t, response.Code, http.StatusOK)
}

func TestSignInUnauthorized(t *testing.T) {
	clearUsersTable()

	setAuthentication()

	req, _ := http.NewRequest("POST", "/signin", nil)

	req.SetBasicAuth("Test User", "wrong-password")

	response := httptest.NewRecorder()
	a.Router.ServeHTTP(response, req)

	actual := string(response.Body.Bytes())

	assert.JSONEq(t, actual, `{"error":"Unauthorized."}`)

	assert.Equal(t, response.Code, http.StatusUnauthorized)

}

func TestSigninAuthorized(t *testing.T) {

	clearUsersTable()

	setAuthentication()
	req, _ := http.NewRequest("POST", "/signin", nil)

	req.SetBasicAuth("Test User", "correct-password")

	response := httptest.NewRecorder()
	a.Router.ServeHTTP(response, req)

	actual := string(response.Body.Bytes())

	assert.JSONEq(t, actual, `{"message":"Sign-in successful."}`)

	assert.Equal(t, response.Code, http.StatusOK)

}

func TestUserInvalidName(t *testing.T) {
	clearUsersTable()

	setAuthentication()
	req, _ := http.NewRequest("POST", "/users", nil)

	req.SetBasicAuth(randSeq(256), "password")

	response := httptest.NewRecorder()
	a.Router.ServeHTTP(response, req)

	actual := string(response.Body.Bytes())

	assert.JSONEq(t, actual, `{"error":"username/password is invalid."}`)

	assert.Equal(t, response.Code, http.StatusBadRequest)

}

func TestCreateUser(t *testing.T) {

	clearUsersTable()

	req, _ := http.NewRequest("POST", "/users", nil)
	req.SetBasicAuth("new-user", "new-password")

	response := httptest.NewRecorder()
	a.Router.ServeHTTP(response, req)

	actual := string(response.Body.Bytes())

	assert.JSONEq(t, actual, `{"id":0,"name":"new-user"}`)

	assert.Equal(t, response.Code, http.StatusOK)
}

func TestOrderIDDoesNotExist(t *testing.T) {
	clearUsersTable()
	clearOrdersTable()

	setAuthentication()
	req, _ := http.NewRequest("GET", "/orders/15?user_id=1", nil)

	req.SetBasicAuth("Test User", "correct-password")

	response := httptest.NewRecorder()
	a.Router.ServeHTTP(response, req)

	actual := string(response.Body.Bytes())

	assert.JSONEq(t, actual, `{"error":"Order not found."}`)

	assert.Equal(t, response.Code, http.StatusNotFound)

}

func TestGetOrder(t *testing.T) {
	clearUsersTable()
	clearOrdersTable()
	clearOrderItemsTable()
	clearItemsTable()

	setAuthentication()

	_, err := a.DB.Exec("INSERT INTO orders(user_id) VALUES('1')")
	if err != nil {
		log.Error(err)
	}

	_, err = a.DB.Exec("INSERT INTO items(name) VALUES('apple')")
	if err != nil {
		log.Error(err)
	}

	_, err = a.DB.Exec("INSERT INTO items(name) VALUES('oranges')")
	if err != nil {
		log.Error(err)
	}

	_, err = a.DB.Exec("INSERT INTO order_items(order_id, item_id) VALUES(1, 1)")
	if err != nil {
		log.Error(err)
	}

	_, err = a.DB.Exec("INSERT INTO order_items(order_id, item_id) VALUES(1, 2)")
	if err != nil {
		log.Error(err)
	}

	req, _ := http.NewRequest("GET", "/orders/1?user_id=1", nil)

	req.SetBasicAuth("Test User", "correct-password")

	response := httptest.NewRecorder()
	a.Router.ServeHTTP(response, req)

	actual := string(response.Body.Bytes())

	assert.JSONEq(t, actual, `{"id":1,"user":"Test User","user_id":0,"items":[{"id":1,"name":"apple"},{"id":2,"name":"oranges"}]}`)

	assert.Equal(t, response.Code, http.StatusOK)
}

func TestGetOrderOfOtherUser(t *testing.T) {
	clearUsersTable()
	clearOrdersTable()
	clearOrderItemsTable()
	clearItemsTable()

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("correct-password"), 8)
	if err != nil {
		log.Error(err)
	}

	statement := fmt.Sprintf(`INSERT INTO users(name,password) VALUES('%s', '%s')`, "First User", hashedPassword)
	_, err = a.DB.Exec(statement)
	if err != nil {
		log.Error(err)
	}

	statement = fmt.Sprintf(`INSERT INTO users(name,password) VALUES('%s', '%s')`, "Attacker User", hashedPassword)
	_, err = a.DB.Exec(statement)
	if err != nil {
		log.Error(err)
	}

	_, err = a.DB.Exec("INSERT INTO orders(user_id) VALUES('1')")
	if err != nil {
		log.Error(err)
	}

	_, err = a.DB.Exec("INSERT INTO items(name) VALUES('apple')")
	if err != nil {
		log.Error(err)
	}

	_, err = a.DB.Exec("INSERT INTO items(name) VALUES('oranges')")
	if err != nil {
		log.Error(err)
	}

	_, err = a.DB.Exec("INSERT INTO order_items(order_id, item_id) VALUES(1, 1)")
	if err != nil {
		log.Error(err)
	}

	_, err = a.DB.Exec("INSERT INTO order_items(order_id, item_id) VALUES(1, 2)")
	if err != nil {
		log.Error(err)
	}
	req, _ := http.NewRequest("GET", "/orders/1?user_id=2", nil)

	req.SetBasicAuth("Attacker User", "correct-password")

	response := httptest.NewRecorder()
	a.Router.ServeHTTP(response, req)

	actual := string(response.Body.Bytes())

	assert.JSONEq(t, actual, `{"error":"Order not found."}`)

	assert.Equal(t, response.Code, http.StatusNotFound)
}

func TestCreateOrder(t *testing.T) {

	clearUsersTable()
	clearOrdersTable()
	clearOrderItemsTable()
	clearItemsTable()

	setAuthentication()

	jsonStr := []byte(`{"user":"Test User", "user_id": 1, "items": [{"id": 1, "name": "Apples"}, {"id": 2, "name": "Oranges"}]}`)

	req, _ := http.NewRequest("POST", "/orders", bytes.NewBuffer(jsonStr))

	req.SetBasicAuth("Test User", "correct-password")

	response := httptest.NewRecorder()
	a.Router.ServeHTTP(response, req)

	actual := string(response.Body.Bytes())

	assert.JSONEq(t, actual, `{"id":1,"user":"Test User","user_id":1,"items":[{"id":1,"name":"Apples"},{"id":2,"name":"Oranges"}]}`)

	assert.Equal(t, response.Code, http.StatusOK)
}

func TestGetOrders(t *testing.T) {
	clearUsersTable()
	clearOrdersTable()
	clearOrderItemsTable()

	setAuthentication()

	_, err := a.DB.Exec("INSERT INTO orders(user_id) VALUES('1')")
	if err != nil {
		log.Error(err)
	}

	_, err = a.DB.Exec("INSERT INTO orders(user_id) VALUES('1')")
	if err != nil {
		log.Error(err)
	}

	_, err = a.DB.Exec("INSERT INTO items(name) VALUES('apple')")
	if err != nil {
		log.Error(err)
	}

	_, err = a.DB.Exec("INSERT INTO items(name) VALUES('oranges')")
	if err != nil {
		log.Error(err)
	}

	_, err = a.DB.Exec("INSERT INTO items(name) VALUES('avacado')")
	if err != nil {
		log.Error(err)
	}

	_, err = a.DB.Exec("INSERT INTO order_items(order_id, item_id) VALUES(1, 1)")
	if err != nil {
		log.Error(err)
	}

	_, err = a.DB.Exec("INSERT INTO order_items(order_id, item_id) VALUES(1, 2)")
	if err != nil {
		log.Error(err)
	}

	_, err = a.DB.Exec("INSERT INTO order_items(order_id, item_id) VALUES(2, 1)")
	if err != nil {
		log.Error(err)
	}

	_, err = a.DB.Exec("INSERT INTO order_items(order_id, item_id) VALUES(2, 3)")
	if err != nil {
		log.Error(err)
	}
	req, _ := http.NewRequest("GET", "/orders?user_id=1&count=10&start=0", nil)

	req.SetBasicAuth("Test User", "correct-password")

	response := httptest.NewRecorder()
	a.Router.ServeHTTP(response, req)

	actual := string(response.Body.Bytes())

	assert.JSONEq(t, actual, `[{"id":2,"user":"Test User","user_id":1,"items":[{"id":1,"name":"apple"},{"id":3,"name":"avacado"}]},{"id":1,"user":"Test User","user_id":1,"items":[{"id":1,"name":"apple"},{"id":2,"name":"oranges"}]}]`)

	assert.Equal(t, response.Code, http.StatusOK)
}

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
