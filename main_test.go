package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/prometheus/common/log"
)

var a App

func TestMain(m *testing.M) {
	a = App{}
	a.InitDB("franklin.db")
	a.InitRouter()

	query :=
		`CREATE TABLE IF NOT EXISTS users (
    id INTEGER AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(255) NOT NULL)`

	if _, err := a.DB.Exec(query); err != nil {
		log.Fatal(err)
	}

	code := m.Run()

	_, err := a.DB.Exec("DELETE FROM users")
	if err != nil {
		log.Error(err)
	}
	_, err = a.DB.Exec("UPDATE SQLITE_SEQUENCE SET SEQ=0 WHERE NAME='users';")
	if err != nil {
		log.Error(err)
	}

	os.Exit(code)
}

func TestUserIDDoesNotExist(t *testing.T) {
	_, err := a.DB.Exec("DELETE FROM users")
	if err != nil {
		log.Error(err)
	}
	_, err = a.DB.Exec("UPDATE SQLITE_SEQUENCE SET SEQ=0 WHERE NAME='users';")
	if err != nil {
		log.Error(err)
	}

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
	_, err := a.DB.Exec("DELETE FROM users")
	if err != nil {
		log.Error(err)
	}

	_, err = a.DB.Exec("UPDATE SQLITE_SEQUENCE SET SEQ=0 WHERE NAME='users';")
	if err != nil {
		log.Error(err)
	}

	_, err = a.DB.Exec("INSERT INTO users(name) VALUES('Test User')")
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

func TestCreateUserInvalidName(t *testing.T) {
	_, err := a.DB.Exec("DELETE FROM users")
	if err != nil {
		log.Error(err)
	}

	_, err = a.DB.Exec("UPDATE SQLITE_SEQUENCE SET SEQ=0 WHERE NAME='users';")
	if err != nil {
		log.Error(err)
	}

	// badName := randSeq(256)

	url := fmt.Sprintf("/user?name=%s", randSeq(256))

	req, _ := http.NewRequest("POST", url, nil)

	response := httptest.NewRecorder()
	a.Router.ServeHTTP(response, req)

	if response.Code != http.StatusBadRequest {
		t.Errorf("Expected response code: %d. Got %d", http.StatusBadRequest, response.Code)
	}

	var m map[string]string
	json.Unmarshal(response.Body.Bytes(), &m)
	if m["error"] != "query parameter is invalid." {
		t.Errorf("Expected the 'error' key of the response to be set to 'query parameter is invalid.'. Got '%s'", m["error"])
	}
}

func TestCreateUser(t *testing.T) {
	_, err := a.DB.Exec("DELETE FROM users")
	if err != nil {
		log.Error(err)
	}

	_, err = a.DB.Exec("UPDATE SQLITE_SEQUENCE SET SEQ=0 WHERE NAME='users';")
	if err != nil {
		log.Error(err)
	}

	req, _ := http.NewRequest("POST", "/user?name=tester", nil)

	response := httptest.NewRecorder()
	a.Router.ServeHTTP(response, req)

	if response.Code != http.StatusOK {
		t.Errorf("Expected response code: %d. Got %d", http.StatusOK, response.Code)
	}
}

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func randSeq(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}
