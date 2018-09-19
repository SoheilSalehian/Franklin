package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
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

func TestSignInUnauthorized(t *testing.T) {
	_, err := a.DB.Exec("DELETE FROM users")
	if err != nil {
		log.Error(err)
	}

	_, err = a.DB.Exec("UPDATE SQLITE_SEQUENCE SET SEQ=0 WHERE NAME='users';")
	if err != nil {
		log.Error(err)
	}

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
	_, err := a.DB.Exec("DELETE FROM users")
	if err != nil {
		log.Error(err)
	}

	_, err = a.DB.Exec("UPDATE SQLITE_SEQUENCE SET SEQ=0 WHERE NAME='users';")
	if err != nil {
		log.Error(err)
	}

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
	_, err := a.DB.Exec("DELETE FROM users")
	if err != nil {
		log.Error(err)
	}

	_, err = a.DB.Exec("UPDATE SQLITE_SEQUENCE SET SEQ=0 WHERE NAME='users';")
	if err != nil {
		log.Error(err)
	}

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
	_, err := a.DB.Exec("DELETE FROM users")
	if err != nil {
		log.Error(err)
	}

	_, err = a.DB.Exec("UPDATE SQLITE_SEQUENCE SET SEQ=0 WHERE NAME='users';")
	if err != nil {
		log.Error(err)
	}

	req, _ := http.NewRequest("POST", "/user/", nil)
	req.SetBasicAuth("new-user", "new-password")

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
