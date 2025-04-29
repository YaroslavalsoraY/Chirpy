package main

import (
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
	platform       string
	secretJWT      string
	polkaKey       string
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

	secretJWT := os.Getenv("SECRET_JWT")

	polkaApi := os.Getenv("POLKA_KEY")

	conf := apiConfig{
		fileserverHits: atomic.Int32{},
		queries:        database.New(db),
		platform:       envPlatform,
		secretJWT:      secretJWT,
		polkaKey:       polkaApi,
	}

	baseHandler := http.FileServer(http.Dir("."))

	mux := http.NewServeMux()
	mux.Handle("/", conf.middlewareMetricsInc(baseHandler))
	mux.Handle("/app/", conf.middlewareMetricsInc(http.StripPrefix("/app", baseHandler)))

	mux.HandleFunc("GET /api/healthz", HandlerHealtzh)
	mux.HandleFunc("GET /admin/metrics", conf.HandlerMetrics)
	mux.HandleFunc("GET /api/chirps", conf.HandlerGetChirps)
	mux.HandleFunc("GET /api/chirps/{chirpID}", conf.HandlerGetOneChirp)

	mux.HandleFunc("POST /admin/reset", conf.HandlerReset)

	mux.HandleFunc("POST /api/chirps", conf.HandlerCreateChirp)
	mux.HandleFunc("POST /api/users", conf.HandlerAddUser)
	mux.HandleFunc("POST /api/login", conf.HandlerLogin)
	mux.HandleFunc("POST /api/refresh", conf.HandlerRefresh)
	mux.HandleFunc("POST /api/revoke", conf.HandlerRevoke)
	mux.HandleFunc("POST /api/polka/webhooks", conf.HandlerPolka)

	mux.HandleFunc("PUT /api/users", conf.HandlerUpdateUser)

	mux.HandleFunc("DELETE /api/chirps/{chirpID}", conf.DeleteChirp)

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
