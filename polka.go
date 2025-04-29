package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/YaroslavalsoraY/Chirpy/internal/auth"
	"github.com/google/uuid"
)

type data struct {
	UserID uuid.UUID `json:"user_id"`
}

type req struct {
	Event string `json:"event"`
	Data  data   `json:"data"`
}

func (cfg *apiConfig) HandlerPolka(w http.ResponseWriter, r *http.Request) {
	keyAPI, err := auth.GetPolkaAPI(r.Header)
	if err != nil || keyAPI != cfg.polkaKey {
		fmt.Println(err)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	decoder := json.NewDecoder(r.Body)
	insertData := req{}
	err = decoder.Decode(&insertData)
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if insertData.Event != "user.upgraded" {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	err = cfg.queries.SetChirpyRed(r.Context(), insertData.Data.UserID)
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
