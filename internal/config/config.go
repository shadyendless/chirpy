package config

import (
	"database/sql"
	"net/http"
	"os"
	"sync/atomic"

	"github.com/shadyendless/chirpy/internal/database"
)

type Config struct {
	DbQueries      *database.Queries
	FileserverHits atomic.Int32
	JWTSecret      string
	Platform       string
}

func New(db *sql.DB) *Config {
	return &Config{
		DbQueries:      database.New(db),
		FileserverHits: atomic.Int32{},
		JWTSecret:      os.Getenv("SECRET"),
		Platform:       os.Getenv("PLATFORM"),
	}
}

func (config *Config) WithMetrics(next http.Handler) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		config.FileserverHits.Add(1)

		next.ServeHTTP(res, req)
	})
}
