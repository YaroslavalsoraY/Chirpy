package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"time"

	"github.com/YaroslavalsoraY/Chirpy/internal/auth"
	"github.com/YaroslavalsoraY/Chirpy/internal/database"
	"github.com/google/uuid"
)

type chirp struct {
	Text string `json:"body"`
}

type returnJson struct {
	InValid   bool      `json:"valid,omitempty"`
	Err       string    `json:"error,omitempty"`
	Body      string    `json:"body,omitempty"`
	CreatedAt time.Time `json:"created_at,omitempty"`
	UpdatedAt time.Time `json:"updated_at,omitempty"`
	ID        uuid.UUID `json:"id,omitempty"`
	UserID    uuid.UUID `json:"user_id,omitempty"`
}

func (cfg *apiConfig) HandlerCreateChirp(w http.ResponseWriter, r *http.Request) {
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	userID, err := auth.ValidateJWT(token, cfg.secretJWT)
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	decoder := json.NewDecoder(r.Body)
	newChirp := chirp{}
	err = decoder.Decode(&newChirp)
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	newChirp.Text = badWordsReplace(newChirp.Text)

	resp := returnJson{
		Err:     "",
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
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if resp.InValid == true {
		w.WriteHeader(http.StatusBadRequest)
		w.Write(respBody)
		return
	}

	arg := database.InsertChirpParams{
		Body:   newChirp.Text,
		UserID: userID,
	}
	returnedChirp, err := cfg.queries.InsertChirp(r.Context(), arg)
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	response := returnJson{
		ID:        returnedChirp.ID,
		CreatedAt: returnedChirp.CreatedAt,
		UpdatedAt: returnedChirp.UpdatedAt,
		Body:      returnedChirp.Body,
		UserID:    returnedChirp.UserID,
	}

	respBody, err = json.Marshal(response)

	w.WriteHeader(http.StatusCreated)
	w.Write(respBody)
	return
}

func (cfg *apiConfig) HandlerGetOneChirp(w http.ResponseWriter, r *http.Request) {
	chirpID, err := uuid.Parse(r.PathValue("chirpID"))
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	chirp, err := cfg.queries.GetOneChirp(r.Context(), chirpID)
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusNotFound)
		return
	}

	respChirp := returnJson{
		ID:        chirp.ID,
		CreatedAt: chirp.CreatedAt,
		UpdatedAt: chirp.UpdatedAt,
		Body:      chirp.Body,
		UserID:    chirp.UserID,
	}

	respData, err := json.Marshal(respChirp)
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Write(respData)
}

func (cfg *apiConfig) HandlerGetChirps(w http.ResponseWriter, r *http.Request) {
	returnChirps := []returnJson{}

	chirps, err := cfg.queries.GetAllChirps(r.Context())
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	authorID := uuid.Nil
	authorIDstring := r.URL.Query().Get("author_id")
	if authorIDstring != "" {
		authorID, err = uuid.Parse(authorIDstring)
		if err != nil {
			fmt.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	sortingMethod := r.URL.Query().Get("sort")
	if sortingMethod == "" {
		sortingMethod = "asc"
	}

	for _, el := range chirps {
		if authorID != uuid.Nil && el.UserID != authorID {
			continue
		}
		returnChirps = append(returnChirps, returnJson{
			ID:        el.ID,
			CreatedAt: el.CreatedAt,
			UpdatedAt: el.UpdatedAt,
			Body:      el.Body,
			UserID:    el.UserID,
		})
	}

	if sortingMethod == "desc" {
		sort.Slice(returnChirps, func(i, j int) bool { return returnChirps[i].CreatedAt.After(returnChirps[j].CreatedAt) })
	}

	resp, err := json.Marshal(returnChirps)
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Write(resp)
}

func (cfg *apiConfig) DeleteChirp(w http.ResponseWriter, r *http.Request) {
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	userID, err := auth.ValidateJWT(token, cfg.secretJWT)
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	chirpID, err := uuid.Parse(r.PathValue("chirpID"))
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	chirp, err := cfg.queries.GetOneChirp(r.Context(), chirpID)
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusNotFound)
		return
	}

	if chirp.UserID != userID {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	err = cfg.queries.DeleteChirp(r.Context(), chirpID)
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
