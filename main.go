package main

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"sync/atomic"

	"github.com/YaroslavalsoraY/Chirpy/internal/database"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type apiConfig struct {
	fileserverHits atomic.Int32
	queries        *database.Queries
	platform		string
}

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		fmt.Println(err)
		return
	}
	dbURL := os.Getenv("DB_URL")
	db, err := sql.Open("postgres", dbURL)

	envPlatform := os.Getenv("PLATFORM")

	conf := apiConfig{
		fileserverHits: atomic.Int32{},
		queries:        database.New(db),
		platform: envPlatform,
	}

	baseHandler := http.FileServer(http.Dir("."))

	mux := http.NewServeMux()
	mux.Handle("/", conf.middlewareMetricsInc(baseHandler))
	mux.Handle("/app/", conf.middlewareMetricsInc(http.StripPrefix("/app", baseHandler)))

	mux.HandleFunc("/api/healthz", HandlerHealtzh)
	mux.HandleFunc("/admin/metrics", conf.HandlerMetrics)
	mux.HandleFunc("/admin/reset", conf.HandlerReset)
	mux.HandleFunc("/api/validate_chirp", HandlerValidate)
	mux.HandleFunc("/api/users", conf.HandlerAddUser)

	server := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	err = server.ListenAndServe()
	if err != nil {
		fmt.Println(err)
		return
	}
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

func (cfg *apiConfig) reset(cont context.Context) {
	cfg.fileserverHits.Swap(0)
	cfg.queries.DeleteAllUsers(cont)
}