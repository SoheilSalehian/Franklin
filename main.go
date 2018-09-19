package main

import (
	"net/http"

	_ "github.com/mattn/go-sqlite3"
	"github.com/prometheus/common/log"
)

func main() {

	a := App{}

	err := a.InitDB("franklin.db")
	if err != nil {
		log.Fatal("Database initialization failed:", err)
	}

	a.InitRouter()

	log.Fatal(http.ListenAndServe(":8080", a.Router))
}
