package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"

	"reviewer-service/internal/db"
	"reviewer-service/internal/handlers"
)

func main() {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://postgres:postgres@db:5432/reviewer?sslmode=disable"
	}

	if err := db.Init(dsn); err != nil {
		log.Fatalf("db init: %v", err)
	}
	defer db.Close()

	r := mux.NewRouter()

	handlers.RegisterTeamRoutes(r)
	handlers.RegisterUserRoutes(r)
	handlers.RegisterPRRoutes(r)

	r.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok"}`))
	})

	srv := &http.Server{
		Addr:              ":8080",
		Handler:           r,
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Println("server listening on :8080")
	if err := srv.ListenAndServe(); err != nil {
		log.Fatalf("server: %v", err)
	}
}
