package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/shadyendless/chirpy/admin"
	"github.com/shadyendless/chirpy/api"
	"github.com/shadyendless/chirpy/config"
)

func main() {
	godotenv.Load()

	serveMux := http.NewServeMux()
	server := http.Server{
		Addr:    ":8080",
		Handler: serveMux,
	}

	dbURL := os.Getenv("DB_URL")
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		fmt.Printf("Could not connect to the database: %v", err)
		os.Exit(1)
	}

	cfg := config.New(db, os.Getenv("PLATFORM"))

	serveMux.Handle("/app/", http.StripPrefix("/app", cfg.WithMetrics(http.FileServer(http.Dir(".")))))

	serveMux.HandleFunc("GET /api/healthz", api.HealthHandler)

	serveMux.HandleFunc("GET /admin/metrics", admin.GetMetricsHandler(cfg))
	serveMux.HandleFunc("POST /admin/reset", admin.ResetUsersHandler(cfg))

	serveMux.HandleFunc("POST /api/validate_chirp", api.ValidateChirpHandler)
	serveMux.HandleFunc("GET /api/chirps/{chirpID}", api.GetChirp(cfg))
	serveMux.HandleFunc("GET /api/chirps", api.GetChirpsHandler(cfg))
	serveMux.HandleFunc("POST /api/chirps", api.CreateChirpHandler(cfg))
	serveMux.HandleFunc("POST /api/users", api.CreateUserHandler(cfg))

	serveMux.HandleFunc("POST /api/login", api.LoginHandler(cfg))

	server.ListenAndServe()
}
