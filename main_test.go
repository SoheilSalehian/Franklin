package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"golang.org/x/crypto/bcrypt"

	"github.com/prometheus/common/log"
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

	req, _ := http.NewRequest("GET", "/user/15", nil)

	response := httptest.NewRecorder()
	a.Router.ServeHTTP(response, req)

	if response.Code != http.StatusNotFound {
		t.Errorf("Expected response code: %d. Got %d", http.StatusNotFound, response.Code)
	}

	var m map[string]string
	json.Unmarshal(response.Body.Bytes(), &m)
	if m["error"] != "User not found" {
		t.Errorf("Expected the 'error' key of the response to be set to 'User not found'. Got '%s'", m["error"])
	}
}

func TestGetUser(t *testing.T) {
	clearUsersTable()

	_, err := a.DB.Exec("INSERT INTO users(name) VALUES('Test User')")
	if err != nil {
		log.Error(err)
	}

	req, _ := http.NewRequest("GET", "/user/1", nil)

	response := httptest.NewRecorder()
	a.Router.ServeHTTP(response, req)

	if response.Code != http.StatusOK {
		t.Errorf("Expected response code: %d. Got %d", http.StatusOK, response.Code)
	}
}

func TestSignInUnauthorized(t *testing.T) {
	clearUsersTable()

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("correct-password"), 8)
	if err != nil {
		log.Error(err)
	}

	statement := fmt.Sprintf(`INSERT INTO users(name,password) VALUES('%s', '%s')`, "Test User", hashedPassword)
	_, err = a.DB.Exec(statement)
	if err != nil {
		log.Error(err)
	}

	req, _ := http.NewRequest("POST", "/signin/", nil)

	req.SetBasicAuth("Test User", "wrong-password")

	response := httptest.NewRecorder()
	a.Router.ServeHTTP(response, req)

	if response.Code != http.StatusUnauthorized {
		t.Errorf("Expected response code: %d. Got %d", http.StatusUnauthorized, response.Code)
	}
}

func TestSigninAuthorized(t *testing.T) {

	clearUsersTable()

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("correct-password"), 8)
	if err != nil {
		log.Error(err)
	}

	statement := fmt.Sprintf(`INSERT INTO users(name,password) VALUES('%s', '%s')`, "Test User", hashedPassword)
	_, err = a.DB.Exec(statement)
	if err != nil {
		log.Error(err)
	}

	req, _ := http.NewRequest("POST", "/signin/", nil)

	req.SetBasicAuth("Test User", "correct-password")

	response := httptest.NewRecorder()
	a.Router.ServeHTTP(response, req)

	if response.Code != http.StatusOK {
		t.Errorf("Expected response code: %d. Got %d", http.StatusOK, response.Code)
	}
}

func TestUserInvalidName(t *testing.T) {
	clearUsersTable()

	req, _ := http.NewRequest("POST", "/user/", nil)

	req.SetBasicAuth(randSeq(256), "password")

	response := httptest.NewRecorder()
	a.Router.ServeHTTP(response, req)

	if response.Code != http.StatusBadRequest {
		t.Errorf("Expected response code: %d. Got %d", http.StatusBadRequest, response.Code)
	}

	var m map[string]string
	json.Unmarshal(response.Body.Bytes(), &m)
	if m["error"] != "username/password is invalid." {
		t.Errorf("Expected the 'error' key of the response to be set to 'username/password is invalid.'. Got '%s'", m["error"])
	}
}

func TestCreateUser(t *testing.T) {

	clearUsersTable()

	req, _ := http.NewRequest("POST", "/user/", nil)
	req.SetBasicAuth("new-user", "new-password")

	response := httptest.NewRecorder()
	a.Router.ServeHTTP(response, req)

	if response.Code != http.StatusOK {
		t.Errorf("Expected response code: %d. Got %d", http.StatusOK, response.Code)
	}
}

func TestOrderIDDoesNotExist(t *testing.T) {
	clearUsersTable()
	clearOrdersTable()

	req, _ := http.NewRequest("GET", "/order/15?user_id=1", nil)

	response := httptest.NewRecorder()
	a.Router.ServeHTTP(response, req)

	if response.Code != http.StatusNotFound {
		t.Errorf("Expected response code: %d. Got %d", http.StatusNotFound, response.Code)
	}

	var m map[string]string
	json.Unmarshal(response.Body.Bytes(), &m)
	if m["error"] != "Order not found" {
		t.Errorf("Expected the 'error' key of the response to be set to 'Order not found'. Got '%s'", m["error"])
	}
}

func TestGetOrder(t *testing.T) {
	clearUsersTable()
	clearOrdersTable()
	clearOrderItemsTable()
	clearItemsTable()

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("correct-password"), 8)
	if err != nil {
		log.Error(err)
	}

	statement := fmt.Sprintf(`INSERT INTO users(name,password) VALUES('%s', '%s')`, "Test User", hashedPassword)
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

	req, _ := http.NewRequest("GET", "/order/1?user_id=1", nil)

	response := httptest.NewRecorder()
	a.Router.ServeHTTP(response, req)

	if response.Code != http.StatusOK {
		t.Errorf("Expected response code: %d. Got %d", http.StatusOK, response.Code)
	}
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
	req, _ := http.NewRequest("GET", "/order/1?user_id=2", nil)

	response := httptest.NewRecorder()
	a.Router.ServeHTTP(response, req)

	if response.Code != http.StatusNotFound {
		t.Errorf("Expected response code: %d. Got %d", http.StatusNotFound, response.Code)
	}
}

func TestCreateOrder(t *testing.T) {

	clearUsersTable()
	clearOrdersTable()
	clearOrderItemsTable()
	clearItemsTable()

	jsonStr := []byte(`{"user":"Test User", "user_id": 1, "items": [{"id": 1, "name": "Apples"}, {"id": 2, "name": "Oranges"}]}`)

	req, _ := http.NewRequest("POST", "/order/", bytes.NewBuffer(jsonStr))

	response := httptest.NewRecorder()
	a.Router.ServeHTTP(response, req)

	if response.Code != http.StatusOK {
		t.Errorf("Expected response code: %d. Got %d", http.StatusOK, response.Code)
	}
}

func TestGetOrders(t *testing.T) {
	clearUsersTable()
	clearOrdersTable()
	clearOrderItemsTable()

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("correct-password"), 8)
	if err != nil {
		log.Error(err)
	}

	statement := fmt.Sprintf(`INSERT INTO users(name,password) VALUES('%s', '%s')`, "Test User", hashedPassword)
	_, err = a.DB.Exec(statement)
	if err != nil {
		log.Error(err)
	}

	_, err = a.DB.Exec("INSERT INTO orders(user_id) VALUES('1')")
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

	req, _ := http.NewRequest("GET", "/orders/?user_id=1&count=10&start=0", nil)

	response := httptest.NewRecorder()
	a.Router.ServeHTTP(response, req)

	if response.Code != http.StatusOK {
		t.Errorf("Expected response code: %d. Got %d", http.StatusOK, response.Code)
	}
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
