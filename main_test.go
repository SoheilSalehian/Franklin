package main

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"golang.org/x/crypto/bcrypt"

	"github.com/jarcoal/httpmock"
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
    password TEXT NOT NULL,
    zip INTEGER,
    store_lat REAL,
    store_lon REAL
)`

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

	assert.JSONEq(t, `{"error":"User not found."}`, actual)

	assert.Equal(t, response.Code, http.StatusNotFound)
}

func TestGetUser(t *testing.T) {
	clearUsersTable()

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("correct-password"), 8)
	if err != nil {
		log.Error(err)
	}

	statement := "INSERT INTO users(name,password,zip,store_lat,store_lon) VALUES(?, ?, ? ,? ,?)"
	_, err = a.DB.Exec(statement, "Test User", hashedPassword, 77777, 22.33, 44.55)
	if err != nil {
		log.Error(err)
	}

	req, _ := http.NewRequest("GET", "/users/1", nil)

	req.SetBasicAuth("Test User", "correct-password")

	response := httptest.NewRecorder()
	a.Router.ServeHTTP(response, req)

	actual := string(response.Body.Bytes())

	assert.JSONEq(t, `{"id":1,"name":"Test User","closest_store":{"coordinates":[22.33,44.55]}}`, actual)

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

	assert.JSONEq(t, `{"error":"Unauthorized."}`, actual)

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

	assert.JSONEq(t, `{"message":"Sign-in successful."}`, actual)

	assert.Equal(t, response.Code, http.StatusOK)

}

func TestUserInvalidName(t *testing.T) {
	clearUsersTable()

	setAuthentication()

	str := fmt.Sprintf(`{"name":"%s", "password": "new-password", "zipcode": 78704}"`, randSeq(256))
	jsonStr := []byte(str)

	req, _ := http.NewRequest("POST", "/users", bytes.NewBuffer(jsonStr))

	response := httptest.NewRecorder()
	a.Router.ServeHTTP(response, req)

	actual := string(response.Body.Bytes())

	assert.JSONEq(t, `{"error":"username/password is invalid."}`, actual)

	assert.Equal(t, response.Code, http.StatusBadRequest)

}

func TestCreateUser(t *testing.T) {

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	fakeResponseJson := []byte(`[
	{
	"no": 1253,
	"name": "Fake Supercenter",
	"country": "US",
	"coordinates": [
	-97.753926,
	30.221033
	],
	"streetAddress": "710 E Ben White Blvd",
	"city": "Austin",
	"stateProvCode": "TX",
	"zip": "78704",
	"phoneNumber": "512-443-6601",
	"sundayOpen": true,
	"timezone": "CST"
	},
	{
	"no": 2133,
	"name": "Fake Supercenter",
	"country": "US",
	"coordinates": [
	-97.8232981,
	30.2322111
	],
	"streetAddress": "5017 W Highway 290",
	"city": "Austin",
	"stateProvCode": "TX",
	"zip": "78735",
	"phoneNumber": "512-892-6086",
	"sundayOpen": true,
	"timezone": "CST"
	}]`)

	url := "http://api.walmartlabs.com/v1/stores"
	httpmock.RegisterResponder("GET", url, httpmock.NewBytesResponder(http.StatusOK, fakeResponseJson))

	clearUsersTable()

	jsonStr := []byte(`{"name":"Test User", "password": "new-password", "zipcode": 78704}`)

	req, _ := http.NewRequest("POST", "/users", bytes.NewBuffer(jsonStr))

	response := httptest.NewRecorder()
	a.Router.ServeHTTP(response, req)

	actual := string(response.Body.Bytes())

	expected := `{"id":0,"name":"Test User","zipcode":78704,"closest_store":{"city":"Austin","coordinates":[-97.753926,30.221033],"country":"US","name":"Fake Supercenter","no":1253,"phoneNumber":"512-443-6601","stateProvCode":"TX","streetAddress":"710 E Ben White Blvd","sundayOpen":true,"timezone":"CST","zip":"78704"}}`

	assert.JSONEq(t, expected, actual)

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

	assert.JSONEq(t, `{"error":"Order not found."}`, actual)

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

	assert.JSONEq(t, `{"id":1,"user":"Test User","user_id":0,"items":[{"id":1,"name":"apple"},{"id":2,"name":"oranges"}]}`, actual)

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

	assert.JSONEq(t, `{"error":"Order not found."}`, actual)

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

	assert.JSONEq(t, `{"id":1,"user":"Test User","user_id":1,"items":[{"id":1,"name":"Apples"},{"id":2,"name":"Oranges"}]}`, actual)

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

	assert.JSONEq(t, `[{"id":2,"user":"Test User","user_id":1,"items":[{"id":1,"name":"apple"},{"id":3,"name":"avacado"}]},{"id":1,"user":"Test User","user_id":1,"items":[{"id":1,"name":"apple"},{"id":2,"name":"oranges"}]}]`, actual)

	assert.Equal(t, response.Code, http.StatusOK)
}

func TestUpdateOrder(t *testing.T) {
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

	jsonStr := []byte(`{"user":"Test User", "user_id": 1, "items": [{"id": 1, "name": "apples"}, {"id": 3, "name": "avacado"}]}`)

	req, _ := http.NewRequest("PUT", "/orders/1", bytes.NewBuffer(jsonStr))

	req.SetBasicAuth("Test User", "correct-password")

	response := httptest.NewRecorder()
	a.Router.ServeHTTP(response, req)

	actual := string(response.Body.Bytes())

	assert.JSONEq(t, `{"id":1,"user":"Test User","user_id":1,"items":[{"id":1,"name":"apples"},{"id":3,"name":"avacado"}]}`, actual)

	assert.Equal(t, response.Code, http.StatusOK)
}

func TestDeleteOrder(t *testing.T) {
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

	_, err = a.DB.Exec("INSERT INTO items(name) VALUES('avacado')")
	if err != nil {
		log.Error(err)
	}

	_, err = a.DB.Exec("INSERT INTO order_items(order_id, item_id) VALUES(1, 1)")
	if err != nil {
		log.Error(err)
	}

	_, err = a.DB.Exec("INSERT INTO order_items(order_id, item_id) VALUES(2, 2)")
	if err != nil {
		log.Error(err)
	}

	jsonStr := []byte(`{"user":"Test User", "user_id": 1}`)

	req, _ := http.NewRequest("DELETE", "/orders/1", bytes.NewBuffer(jsonStr))

	req.SetBasicAuth("Test User", "correct-password")

	response := httptest.NewRecorder()
	a.Router.ServeHTTP(response, req)

	actual := string(response.Body.Bytes())
	log.Info(actual)

	assert.JSONEq(t, `{"id":1,"user":"Test User","user_id":1,"items":null}`, actual)

	assert.Equal(t, response.Code, http.StatusOK)
}

func TestDeleteOtherUsersOrder(t *testing.T) {
	clearUsersTable()
	clearOrdersTable()
	clearOrderItemsTable()
	clearItemsTable()

	setAuthentication()

	_, err := a.DB.Exec("INSERT INTO orders(id, user_id) VALUES('1', '1')")
	if err != nil {
		log.Error(err)
	}

	_, err = a.DB.Exec("INSERT INTO orders(id, user_id) VALUES('2', '2')")
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

	_, err = a.DB.Exec("INSERT INTO order_items(order_id, item_id) VALUES(1, 3)")
	if err != nil {
		log.Error(err)
	}

	jsonStr := []byte(`{"user":"First User", "user_id": 1}`)

	req, _ := http.NewRequest("DELETE", "/orders/1", bytes.NewBuffer(jsonStr))

	req.SetBasicAuth("Test User", "correct-password")

	response := httptest.NewRecorder()
	a.Router.ServeHTTP(response, req)

	actual := string(response.Body.Bytes())

	assert.JSONEq(t, `{"error":"Forbidden."}`, actual)

	assert.Equal(t, response.Code, http.StatusForbidden)
}
