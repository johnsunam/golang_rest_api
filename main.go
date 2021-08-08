package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

func main() {
	db, err := OpenConnection()
	if err != nil {
		fmt.Printf("\n error %+v\n", err)
	}
	app := &App{
		Router:   mux.NewRouter().StrictSlash(true),
		Database: db,
	}

	app.SetupRouter()

	log.Fatal(http.ListenAndServe(":8080", app.Router))
}
