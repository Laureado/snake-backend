package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"main/logs"
	"main/models"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/cors"
	_ "github.com/lib/pq"
)

func Routes() *chi.Mux {
	mux := chi.NewMux()

	// global middlewares
	mux.Use(
		cors.Handler(cors.Options{
			AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
			AllowedHeaders: []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		}),
		middleware.Logger,    //log every http request
		middleware.Recoverer, //recover if a panic occurs

	)

	mux.Post("/ranking", SaveRankingHandler)
	mux.Get("/ranking", ShowRankingHandler)

	return mux
}

// ranking POST
func SaveRankingHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	cmd := parseRequest(r)

	// Validate
	if len(cmd.Name) > 20 || len(cmd.Name) < 1 {
		w.WriteHeader(http.StatusBadRequest)
		logs.Log().Error("cannot execute statement")
		m := map[string]interface{}{"msg": "name must have at least 1 character and less than 21"}
		_ = json.NewEncoder(w).Encode(m)
		return
	}
	if cmd.Points < 1 {
		w.WriteHeader(http.StatusBadRequest)
		logs.Log().Error("cannot execute statement")
		m := map[string]interface{}{"msg": "points must be at least 1"}
		_ = json.NewEncoder(w).Encode(m)
		return
	}

	// connect to db
	db, err := sql.Open("postgres", "postgresql://jesus@localhost:20257/ranking_db?sslmode=disable")
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		logs.Log().Error("cannot execute statement")
		m := map[string]interface{}{"msg": "error in create rank"}
		_ = json.NewEncoder(w).Encode(m)
		return
	}

	// insert
	lastInsertId := ""
	err = db.QueryRow("INSERT INTO ranking (Name, Points) VALUES ($1, $2) RETURNING id", cmd.Name, cmd.Points).Scan(&lastInsertId)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		logs.Log().Error("cannot execute statement")
		m := map[string]interface{}{"msg": "error in create rank"}
		_ = json.NewEncoder(w).Encode(m)
		return
	}

	res := map[string]interface{}{"msg": "ok"}
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(res)

	_ = db.Close()

}

func parseRequest(r *http.Request) *models.CreateRankingCMD {
	body := r.Body

	defer body.Close()
	var cmd models.CreateRankingCMD

	_ = json.NewDecoder(body).Decode(&cmd)

	return &cmd
}

// ranking get
func ShowRankingHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	db, err := sql.Open("postgres", "postgresql://jesus@localhost:20257/ranking_db?sslmode=disable")
	if err != nil {
		logs.Log().Error("cannot create transaction")
	}

	rows, err := db.Query("SELECT Id, Name, Points FROM ranking ORDER BY Points desc limit 10")
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		logs.Log().Error("cannot execute statement")
		m := map[string]interface{}{"msg": "error in getting rank"}
		_ = json.NewEncoder(w).Encode(m)
		return
	}
	defer rows.Close()

	ranks := []*models.CreateRankingCMD{}

	for rows.Next() {
		var Id, Name string
		var Points int
		if err := rows.Scan(&Id, &Name, &Points); err != nil {
			log.Fatal(err)
		}
		ranks = append(ranks, &models.CreateRankingCMD{Id: Id, Name: Name, Points: Points})
	}

	_ = json.NewEncoder(w).Encode(ranks)
	_ = db.Close()
}
