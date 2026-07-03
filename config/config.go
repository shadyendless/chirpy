package config

import (
	"database/sql"
	"net/http"
	"sync/atomic"

	"github.com/shadyendless/chirpy/internal/database"
)

type Config struct {
	DbQueries      *database.Queries
	FileserverHits atomic.Int32
	Platform       string
}

func New(db *sql.DB, platform string) *Config {
	return &Config{
		DbQueries:      database.New(db),
		FileserverHits: atomic.Int32{},
		Platform:       platform,
	}
}

func (config *Config) WithMetrics(next http.Handler) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		config.FileserverHits.Add(1)

		next.ServeHTTP(res, req)
	})
}
