package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func HandlerHealtzh(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func (cfg *apiConfig) HandlerMetrics(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	w.Header().Add("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	count := fmt.Sprintf("<html><body><h1>Welcome, Chirpy Admin</h1><p>Chirpy has been visited %d times!</p></body></html>", cfg.fileserverHits.Load())
	w.Write([]byte(count))
}

func (cfg *apiConfig) HandlerReset(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	if cfg.platform != "dev" {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	cfg.reset(r.Context())
	w.WriteHeader(http.StatusOK)
}

func HandlerValidate(w http.ResponseWriter, r *http.Request) {
	type chirpText struct {
		Text string `json:"body"`
	}

	type returnJson struct {
		Err   string `json:"error,omitempty"`
		Valid bool   `json:"valid"`
	}

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	decoder := json.NewDecoder(r.Body)
	message := chirpText{}
	err := decoder.Decode(&message)
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(500)
		return
	}

	resp := returnJson{
		Err:   "",
		Valid: true,
	}

	w.Header().Set("Content-Type", "application/json")

	if len(message.Text) > 140 {
		resp.Err = "Chirp is too long"
		resp.Valid = false
	}

	respBody, err := json.Marshal(resp)
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(500)
		return
	}

	if resp.Valid == false {
		w.WriteHeader(http.StatusBadRequest)
		w.Write(respBody)
		return
	}

	w.Write(respBody)
	return
}

func (cfg *apiConfig) HandlerAddUser(w http.ResponseWriter, r *http.Request) {
	type info struct {
		Email string `json:"email"`
	}

	if r.Method != http.MethodPost {
		w.WriteHeader(500)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	decoder := json.NewDecoder(r.Body)
	user := info{}
	err := decoder.Decode(&user)
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(500)
		return
	}

	newUser, err := cfg.queries.CreateUser(r.Context(), user.Email)
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(500)
		return
	}

	resp, err := json.Marshal(newUser)
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(500)
		return
	}

	w.WriteHeader(201)
	w.Write(resp)
	return
}
