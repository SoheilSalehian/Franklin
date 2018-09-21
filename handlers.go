package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
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
	log.Info("running on port :8080")
	return nil
}

func (a *App) InitRouter() {
	a.Router = mux.NewRouter()

	a.Router.HandleFunc("/users/{id:[0-9]+}", a.basicAuth(a.getUser)).Methods("GET")
	a.Router.HandleFunc("/users", a.createUser).Methods("POST")

	a.Router.HandleFunc("/orders/{id:[0-9]+}", a.basicAuth(a.getOrder)).Queries("user_id", "{user_id}").Methods("GET")
	a.Router.HandleFunc("/orders", a.basicAuth(a.getOrders)).Queries("user_id", "{user_id}").Methods("GET")
	a.Router.HandleFunc("/orders", a.basicAuth(a.createOrder)).Methods("POST")
	a.Router.HandleFunc("/orders/{id:[0-9]+}", a.basicAuth(a.updateOrder)).Methods("PUT")
	a.Router.HandleFunc("/orders/{id:[0-9]+}", a.basicAuth(a.deleteOrder)).Methods("DELETE")

	a.Router.HandleFunc("/signin", a.basicAuth(a.signin)).Methods("POST")
}

// User handlers
//
//

// TODO: need to add email verification with redirct to secure this
func (a *App) createUser(w http.ResponseWriter, r *http.Request) {
	u := User{}

	err := json.NewDecoder(r.Body).Decode(&u)
	if err != nil {
		log.Error(err)
		respondWithError(w, http.StatusBadRequest, "Order ID is invalid.")
		return
	}

	if !userValidations(u.Name, u.Password) {
		log.Error("User name validation failed.")
		respondWithError(w, http.StatusBadRequest, "username/password is invalid.")
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(u.Password), 8)
	if err != nil {
		log.Error(err)
		respondWithError(w, http.StatusInternalServerError, "User could not be created.")
	}

	u.Password = string(hashedPassword)

	u.ClosestStore, err = a.storeLocator(u.Zipcode)
	if err != nil {
		log.Error(err)
		respondWithError(w, http.StatusInternalServerError, "User could not be created.")
		return
	}

	if err := u.createUser(a.DB); err != nil {
		log.Error(err)
		respondWithError(w, http.StatusInternalServerError, "User could not be created.")
		return
	}

	respondWithJSON(w, http.StatusOK, u)
}

func (a *App) storeLocator(zipcode int) (Store, error) {

	apiKey := os.Getenv("WALMART_OPEN_API_KEY")
	if apiKey == "" {
		log.Error("WALMART_OPEN_API_KEY is not set.")
	}

	url := fmt.Sprintf("http://api.walmartlabs.com/v1/stores?apiKey=%s&zip=%s&format=json", apiKey, strconv.Itoa(zipcode))

	resp, err := http.Get(url)
	if err != nil {
		log.Error(err)
		return Store{}, err
	}
	defer resp.Body.Close()

	var stores []Store
	err = json.NewDecoder(resp.Body).Decode(&stores)
	if err != nil {
		log.Error(err)
		return Store{}, err
	}

	// TODO: Make this more intelligent (geo-location/order inventory based)
	return stores[0], nil

}

// FIXME: securing this based on username
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
		respondWithError(w, http.StatusNotFound, "User not found.")
		return
	}
	respondWithJSON(w, http.StatusOK, u)
}

// Order handlers
//
//

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
		respondWithError(w, http.StatusNotFound, "Order not found.")
		return
	}
	respondWithJSON(w, http.StatusOK, o)
}

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

func (a *App) getOrders(w http.ResponseWriter, r *http.Request) {
	userID := r.FormValue("user_id")
	v := r.URL.Query()
	count, _ := strconv.Atoi(v.Get("count"))
	start, _ := strconv.Atoi(v.Get("start"))

	// TODO: add proper valiations
	if count > 10 || count < 1 {
		count = 10
	}
	if start < 0 {
		start = 0
	}

	orders, err := getOrders(a.DB, userID, count, start)
	if err != nil {
		log.Error(err)
		respondWithError(w, http.StatusNotFound, "No orders found.")
		return
	}

	respondWithJSON(w, http.StatusOK, orders)
}

func (a *App) updateOrder(w http.ResponseWriter, r *http.Request) {

	o := Order{}

	vars := mux.Vars(r)
	user, _, _ := r.BasicAuth()

	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		log.Error(err)
		respondWithError(w, http.StatusBadRequest, "User ID is invalid.")
		return
	}

	o.ID = id

	err = json.NewDecoder(r.Body).Decode(&o)
	if err != nil {
		log.Error(err)
		respondWithError(w, http.StatusBadRequest, "Order ID is invalid.")
		return
	}

	if user != o.User {
		log.Error("Unathorized attempt to update order by user: ", user)
		respondWithError(w, http.StatusForbidden, "Forbidden.")
		return
	}

	if err := o.updateOrder(a.DB); err != nil {
		log.Error(err)
		if err.Error() == "Order not found." {
			respondWithError(w, http.StatusNotFound, "Order could not be found.")
		} else {
			respondWithError(w, http.StatusInternalServerError, "Order could not be updated.")
		}
		return
	}
	respondWithJSON(w, http.StatusOK, o)
}

func (a *App) deleteOrder(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)

	user, _, _ := r.BasicAuth()

	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		log.Error(err)
		respondWithError(w, http.StatusBadRequest, "User ID is invalid.")
		return
	}

	o := Order{}
	o.ID = id

	err = json.NewDecoder(r.Body).Decode(&o)
	if err != nil {
		log.Error(err)
		respondWithError(w, http.StatusBadRequest, "Order ID is invalid.")
		return
	}

	if user != o.User {
		log.Error("Unathorized attempt to delete order by user: ", user)
		respondWithError(w, http.StatusForbidden, "Forbidden.")
		return
	}

	if err := o.deleteOrder(a.DB); err != nil {
		log.Error(err)
		if err.Error() == "Order doesn't exist." {
			respondWithError(w, http.StatusNotFound, "Order doesn't exist.")
			return
		} else {
			respondWithError(w, http.StatusInternalServerError, "Order could not be deleted.")
			return
		}
	}
	respondWithJSON(w, http.StatusOK, o)
}

// Genearal handlers and middleware
//
//

func (a *App) signin(w http.ResponseWriter, r *http.Request) {
	respondWithJSON(w, http.StatusOK, map[string]string{"message": "Sign-in successful."})
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
