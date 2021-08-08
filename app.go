package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

var Iso8601 = "2006-01-02T15:04:05Z"

type App struct {
	Router   *mux.Router
	Database *sql.DB
}

type PlaceRequest struct {
	Name string `json:"name"`
	Code string `json:"code"`
}
type Place struct {
	Id        string `json:"id"`
	Name      string `json:"name"`
	Code      string `json:"code"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}
type HTTPErrors struct {
	Cause  error  `json:"-"`
	Detail string `json:"detail"`
	Status int    `json:"-"`
}

func (app *App) SetupRouter() {
	app.Router.
		Methods("GET").
		Path("/places").
		HandlerFunc(app.listPlace)

	app.Router.
		Methods("GET").
		Path("/places/{id}").
		HandlerFunc(app.getPlaceById)

	app.Router.
		Methods("PUT").
		Path("/places/{id}").
		HandlerFunc(app.updatePlaceById)

	app.Router.
		Methods("POST").
		Path("/places").
		HandlerFunc(app.createPlace)

}

func (app *App) getPlaceById(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, ok := vars["id"]
	if !ok {
		log.Fatal("No ID in the path")
	}
	dbdata := &Place{}
	row := app.Database.QueryRow("SELECT * FROM places WHERE id=$1", id)
	if err := row.Scan(&dbdata.Id, &dbdata.Name, &dbdata.Code, &dbdata.CreatedAt, &dbdata.UpdatedAt); err != nil {
		log.Printf("failed to scan row: %+v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-type", "application/json")
	if err := json.NewEncoder(w).Encode(dbdata); err != nil {
		panic(err)
	}
}

func (app *App) createPlace(w http.ResponseWriter, r *http.Request) {
	var p PlaceRequest
	err := json.NewDecoder(r.Body).Decode(&p)
	if err != nil {
		log.Printf("failed to decode request body %+v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	createdAt := time.Now().Format(Iso8601)
	sqlQuery := `INSERT INTO places (name, code, created_at, updated_at) VALUES ($1, $2, $3, $4)`
	_, err = app.Database.Exec(sqlQuery, p.Name, p.Code, createdAt, createdAt)
	if err != nil {
		log.Printf("failed to create new place %+v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	json.NewEncoder(w).Encode("Place created successfully.")
}

func (app *App) updatePlaceById(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, ok := vars["id"]
	if !ok {
		log.Fatal("No ID in the path")
	}
	var p PlaceRequest
	err := json.NewDecoder(r.Body).Decode(&p)
	if err != nil {
		log.Printf("failed to decode request body %+v", err)
		err = fmt.Errorf("invalid body parameter: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	dbdata := &Place{}
	updatedAt := time.Now().Format(Iso8601)
	sqlQuery := `UPDATE  places set name=$1, code=$2, updated_at= $3 where id=$4 RETURNING id, name, code, created_at, updated_at`

	row := app.Database.QueryRow(sqlQuery, p.Name, p.Code, updatedAt, id)
	if err := row.Scan(&dbdata.Id, &dbdata.Name, &dbdata.Code, &dbdata.CreatedAt, &dbdata.UpdatedAt); err != nil {
		log.Printf("failed to scan row: %+v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-type", "application/json")
	json.NewEncoder(w).Encode(dbdata)
}

func (app *App) listPlace(w http.ResponseWriter, r *http.Request) {

	rows, err := app.Database.Query("SELECT * FROM places")
	if err != nil {
		log.Printf("failed fetch list of places: %+v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var places []Place
	for rows.Next() {
		var plc Place
		rows.Scan(&plc.Id, &plc.Name, &plc.Code, &plc.CreatedAt, &plc.UpdatedAt)
		places = append(places, plc)
	}

	w.Header().Set("Content-type", "application/json")
	if err := json.NewEncoder(w).Encode(places); err != nil {
		panic(err)
	}
}
