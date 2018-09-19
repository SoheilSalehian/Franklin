package main

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/prometheus/common/log"
	"golang.org/x/crypto/bcrypt"
)

type App struct {
	Router *mux.Router
	DB     *sql.DB
}

func (a *App) InitDB(dbName string) error {
	var err error
	a.DB, err = sql.Open("sqlite3", dbName)
	if err != nil {
		log.Error("can not open connection to sqlite3 db: ", dbName)
		return err
	}

	err = a.DB.Ping()
	if err != nil {
		log.Error("not connected to", dbName)
		return err
	}

	log.Info("successful connection to DB: ", dbName)
	return nil
}

func (a *App) InitRouter() {
	a.Router = mux.NewRouter()
	a.Router.HandleFunc("/signin/", a.basicAuth(a.signin)).Methods("POST")
	a.Router.HandleFunc("/user/", a.createUser).Methods("POST")
	// Built-in id validations
	a.Router.HandleFunc("/user/{id:[0-9]+}", a.updateUser).Methods("PUT")
	a.Router.HandleFunc("/user/{id:[0-9]+}", a.getUser).Methods("GET")
	a.Router.HandleFunc("/user/{id:[0-9]+}", a.deleteUser).Methods("DELETE")

	a.Router.HandleFunc("/order/{id:[0-9]+}", a.getOrder).Queries("user_id", "{user_id}").Methods("GET")
	a.Router.HandleFunc("/order/", a.createOrder).Methods("POST")
}

// User handlers
//
//

// TODO: need to add email verification with redirct to secure this
func (a *App) createUser(w http.ResponseWriter, r *http.Request) {

	username, password, _ := r.BasicAuth()
	if !userValidations(username, password) {
		log.Error("User name validation failed.")
		respondWithError(w, http.StatusBadRequest, "username/password is invalid.")
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 8)
	if err != nil {
		log.Error(err)
		respondWithError(w, http.StatusInternalServerError, "User could not be created.")
	}

	u := User{Name: username, Password: string(hashedPassword)}

	if err := u.createUser(a.DB); err != nil {
		log.Error(err)
		respondWithError(w, http.StatusInternalServerError, "User could not be created.")
		return
	}
	respondWithJSON(w, http.StatusOK, u)
}

// FIXME: need to wrap basicAuth
func (a *App) getUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		log.Error(err)
		respondWithError(w, http.StatusBadRequest, "User ID is invalid.")
		return
	}

	u := User{ID: id}
	if err := u.getUser(a.DB); err != nil {
		log.Error(err)
		respondWithError(w, http.StatusNotFound, "User not found")
		return
	}
	respondWithJSON(w, http.StatusOK, u)
}

func (a *App) deleteUser(w http.ResponseWriter, r *http.Request) {
	respondWithError(w, http.StatusNotImplemented, "n/a")
}

func (a *App) updateUser(w http.ResponseWriter, r *http.Request) {
	respondWithError(w, http.StatusNotImplemented, "n/a")
}

// Order handlers
//
//

// FIXME: need to wrap basicAuth
func (a *App) getOrder(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	userID := r.FormValue("user_id")

	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		log.Error(err)
		respondWithError(w, http.StatusBadRequest, "Order ID is invalid.")
		return
	}

	o := Order{ID: id}
	if err := o.getOrder(a.DB, userID); err != nil {
		log.Error(err)
		respondWithError(w, http.StatusNotFound, "Order not found")
		return
	}
	respondWithJSON(w, http.StatusOK, o)
}

// FIXME: need to wrap basicAuth
func (a *App) createOrder(w http.ResponseWriter, r *http.Request) {
	o := Order{}
	err := json.NewDecoder(r.Body).Decode(&o)
	if err != nil {
		log.Error(err)
		respondWithError(w, http.StatusBadRequest, "Order ID is invalid.")
		return
	}

	if err := o.createOrder(a.DB); err != nil {
		log.Error(err)
		respondWithError(w, http.StatusInternalServerError, "order could not be created.")
		return
	}
	respondWithJSON(w, http.StatusOK, o)
}

// Genearal handlers
//
//

func (a *App) signin(w http.ResponseWriter, r *http.Request) {
	respondWithJSON(w, http.StatusOK, map[string]string{"result": "Sign-in successful."})
}

func (a *App) basicAuth(fn http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		username, password, _ := r.BasicAuth()

		if !userValidations(username, password) {
			log.Error("User name validation failed.")
			respondWithError(w, http.StatusBadRequest, "username/password is invalid.")
			return
		}
		result := a.DB.QueryRow("SELECT password FROM users WHERE name=$1", username)
		if result == nil {
			log.Error("could not find user: ", username)
			respondWithError(w, http.StatusUnauthorized, "Unauthorized.")
			return
		}

		var hashedPassword string

		err := result.Scan(&hashedPassword)
		if err != nil {
			if err == sql.ErrNoRows {
				log.Error(err)
				respondWithError(w, http.StatusUnauthorized, "Unauthorized.")
				return
			}
			log.Error(err)
			respondWithError(w, http.StatusInternalServerError, "Internal server errror.")
			return
		}

		if err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password)); err != nil {
			log.Error(err)
			respondWithError(w, http.StatusUnauthorized, "Unauthorized.")
			return
		}
		fn(w, r)
	}
}

func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, map[string]string{"error": message})
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}
