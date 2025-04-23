package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/YaroslavalsoraY/Chirpy/internal/database"
	"github.com/google/uuid"
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

func (cfg *apiConfig) HandlerCreateChirp(w http.ResponseWriter, r *http.Request) {
	type chirp struct {
		Text string `json:"body"`
		UserID uuid.UUID `json:"user_id"`
	}

	type returnJson struct {
		InValid bool   `json:"valid,omitempty"`
		Err   string `json:"error,omitempty"`
		Body string `json:"body,omitempty"`
		CreatedAt time.Time `json:"created_at,omitempty"`
		UpdatedAt time.Time `json:"updated_at,omitempty"`
		ID    uuid.UUID    `json:"id,omitempty"`
		UserID uuid.UUID `json:"user_id,omitempty"`
	}

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	decoder := json.NewDecoder(r.Body)
	newChirp := chirp{}
	err := decoder.Decode(&newChirp)
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(500)
		return
	}

	resp := returnJson{
		Err:   "",
		InValid: false,
	}

	w.Header().Set("Content-Type", "application/json")

	if len(newChirp.Text) > 140 {
		resp.Err = "Chirp is too long"
		resp.InValid = true
	}

	respBody, err := json.Marshal(resp)
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(500)
		return
	}

	if resp.InValid == true {
		w.WriteHeader(http.StatusBadRequest)
		w.Write(respBody)
		return
	}

	arg := database.InsertChirpParams{
		Body: newChirp.Text,
		UserID: newChirp.UserID,
	}
	returnedChirp, err := cfg.queries.InsertChirp(r.Context(), arg)
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(500)
		return
	}

	response := returnJson{
		ID: returnedChirp.ID,
		CreatedAt: returnedChirp.CreatedAt,
		UpdatedAt: returnedChirp.UpdatedAt,
		Body: returnedChirp.Body,
		UserID: returnedChirp.UserID,
	}

	respBody, err = json.Marshal(response)

	w.WriteHeader(http.StatusCreated)
	w.Write(respBody)
	return
}

func (cfg *apiConfig) HandlerAddUser(w http.ResponseWriter, r *http.Request) {
	type info struct {
		Email string `json:"email"`
	}

	type User struct {
		ID        uuid.UUID `json:"id"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		Email     string    `json:"email"`
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

	respUser := User{
		ID: newUser.ID,
		CreatedAt: newUser.CreatedAt.Time,
		UpdatedAt: newUser.UpdatedAt.Time,
		Email: newUser.Email,
	}

	resp, err := json.Marshal(respUser)
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(500)
		return
	}

	w.WriteHeader(201)
	w.Write(resp)
	return
}
