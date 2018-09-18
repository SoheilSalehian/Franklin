package main

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"

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

	log.Info("connection to DB successful.")
	return nil
}

func (a *App) InitRouter() {
	a.Router = mux.NewRouter()
	a.Router.HandleFunc("/user", a.createUser).Queries("name", "{name}").Methods("POST")
	// Built-in id validations
	a.Router.HandleFunc("/user/{id:[0-9]+}", a.updateUser).Methods("PUT")
	a.Router.HandleFunc("/user/{id:[0-9]+}", a.getUser).Methods("GET")
	a.Router.HandleFunc("/user/{id:[0-9]+}", a.deleteUser).Methods("DELETE")
}

func (a *App) createUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	// FIXME: need proper validation here
	if len(vars["name"]) >= 255 {
		respondWithError(w, http.StatusBadRequest, "query parameter is invalid.")
		return

	}
	u := user{Name: vars["name"]}

	if err := u.createUser(a.DB); err != nil {
		log.Error(err)
		respondWithError(w, http.StatusInternalServerError, "User could not be created.")
		return
	}
	respondWithJSON(w, http.StatusOK, u)
}

func (a *App) updateUser(w http.ResponseWriter, r *http.Request) {
	respondWithError(w, http.StatusNotImplemented, "n/a")
}

func (a *App) getUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
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

func (a *App) deleteUser(w http.ResponseWriter, r *http.Request) {
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

func main() {

	a := App{}

	err := a.InitDB("franklin.db")
	if err != nil {
		log.Fatal("Database initialization failed:", err)
	}

	a.InitRouter()

	log.Fatal(http.ListenAndServe(":8080", a.Router))
}
