package main

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"

	"golang.org/x/crypto/bcrypt"

	"github.com/gorilla/mux"
	_ "github.com/mattn/go-sqlite3"
	"github.com/prometheus/common/log"
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
}

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

	u := user{Name: username, Password: string(hashedPassword)}

	if err := u.createUser(a.DB); err != nil {
		log.Error(err)
		respondWithError(w, http.StatusInternalServerError, "User could not be created.")
		return
	}
	respondWithJSON(w, http.StatusOK, u)
}

func (a *App) getUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		log.Error(err)
		respondWithError(w, http.StatusBadRequest, "User ID is invalid.")
		return
	}

	u := user{ID: id}
	if err := u.getUser(a.DB); err != nil {
		log.Error(err)
		respondWithError(w, http.StatusNotFound, "User not found")
		return
	}
	respondWithJSON(w, http.StatusOK, u)
}

func (a *App) signin(w http.ResponseWriter, r *http.Request) {
	respondWithJSON(w, http.StatusOK, map[string]string{"result": "Sign-in successful."})
}

func (a *App) deleteUser(w http.ResponseWriter, r *http.Request) {
	respondWithError(w, http.StatusNotImplemented, "n/a")
}

func (a *App) updateUser(w http.ResponseWriter, r *http.Request) {
	respondWithError(w, http.StatusNotImplemented, "n/a")
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

// FIXME: need more robust validations (github.com/asaskevich/govalidator)
func userValidations(username string, password string) bool {
	if len(username) >= 255 {
		return false
	}
	return true

}

func main() {

	a := App{}

	err := a.InitDB("franklin.db")
	if err != nil {
		log.Fatal("Database initialization failed:", err)
	}

	a.InitRouter()

	log.Fatal(http.ListenAndServe(":8080", a.Router))
}
