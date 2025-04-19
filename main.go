package main

import (
	"fmt"
	"net/http"
	"sync/atomic"
)

type apiConfig struct {
	fileserverHits atomic.Int32
}

func main() {
	conf := apiConfig {
		fileserverHits: atomic.Int32{},
	}
	
	baseHandler := http.FileServer(http.Dir("."))

	mux := http.NewServeMux()
	mux.Handle("/", conf.middlewareMetricsInc(baseHandler))
	mux.Handle("/app/", conf.middlewareMetricsInc(http.StripPrefix("/app", baseHandler)))
	
	mux.HandleFunc("/api/healthz", HandlerHealtzh)
	mux.HandleFunc("/admin/metrics", conf.HandlerMetrics)
	mux.HandleFunc("/admin/reset", conf.HandlerReset)

	server := &http.Server{
		Addr: ":8080",
		Handler: mux,
	}

	err := server.ListenAndServe()
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

func (cfg *apiConfig) reset() {
	cfg.fileserverHits.Swap(0)
}
