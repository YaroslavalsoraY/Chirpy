package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/YaroslavalsoraY/Chirpy/internal/auth"
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
		w.WriteHeader(500)
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
		w.WriteHeader(500)
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
		w.WriteHeader(500)
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
	type returnChirp struct {
		ID        uuid.UUID `json:"id"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		Body      string    `json:"body"`
		UserID    uuid.UUID `json:"user_id"`
	}

	chirpID, err := uuid.Parse(r.PathValue("chirpID"))
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(500)
		return
	}

	chirp, err := cfg.queries.GetOneChirp(r.Context(), chirpID)
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusNotFound)
		return
	}

	respChirp := returnChirp{
		ID:        chirp.ID,
		CreatedAt: chirp.CreatedAt,
		UpdatedAt: chirp.UpdatedAt,
		Body:      chirp.Body,
		UserID:    chirp.UserID,
	}

	respData, err := json.Marshal(respChirp)
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(500)
		return
	}

	w.Write(respData)
}

func (cfg *apiConfig) HandlerGetChirps(w http.ResponseWriter, r *http.Request) {
	type returnChirp struct {
		ID        uuid.UUID `json:"id"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		Body      string    `json:"body"`
		UserID    uuid.UUID `json:"user_id"`
	}

	returnChirps := []returnChirp{}

	chirps, err := cfg.queries.GetAllChirps(r.Context())
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(500)
		return
	}
	for _, el := range chirps {
		returnChirps = append(returnChirps, returnChirp{
			ID:        el.ID,
			CreatedAt: el.CreatedAt,
			UpdatedAt: el.UpdatedAt,
			Body:      el.Body,
			UserID:    el.UserID,
		})
	}

	resp, err := json.Marshal(returnChirps)
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(500)
		return
	}

	w.Write(resp)
}

func (cfg *apiConfig) HandlerAddUser(w http.ResponseWriter, r *http.Request) {
	type info struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	type User struct {
		ID        	uuid.UUID `json:"id"`
		CreatedAt 	time.Time `json:"created_at"`
		UpdatedAt 	time.Time `json:"updated_at"`
		Email     	string    `json:"email"`
		IsChirpyRed bool      `json:"is_chirpy_red"`
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

	hash, err := auth.HashPassword(user.Password)
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(500)
		return
	}

	args := database.CreateUserParams{
		Email:          user.Email,
		HashedPassword: hash,
	}

	newUser, err := cfg.queries.CreateUser(r.Context(), args)
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(500)
		return
	}

	respUser := User{
		ID:        newUser.ID,
		CreatedAt: newUser.CreatedAt.Time,
		UpdatedAt: newUser.UpdatedAt.Time,
		Email:     newUser.Email,
		IsChirpyRed: false,
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

func (cfg *apiConfig) HandlerLogin(w http.ResponseWriter, r *http.Request) {
	type insertData struct {
		Password string `json:"password"`
		Email    string `json:"email"`
		Expires  int
	}

	type User struct {
		ID           uuid.UUID `json:"id"`
		CreatedAt    time.Time `json:"created_at"`
		UpdatedAt    time.Time `json:"updated_at"`
		Email        string    `json:"email"`
		Token        string    `json:"token"`
		RefreshToken string    `json:"refresh_token"`
		IsChirpyRed  bool  	   `json:"is_chirpy_red"`
	}

	decoder := json.NewDecoder(r.Body)
	authData := insertData{
		Expires: 3600,
	}
	err := decoder.Decode(&authData)
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(500)
		return
	}

	userInfo, err := cfg.queries.GetUserHashedPassword(r.Context(), authData.Email)
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	err = auth.CheckPasswordHash(userInfo.HashedPassword, authData.Password)
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	token, err := auth.MakeJWT(userInfo.ID, cfg.secretJWT, time.Second*time.Duration(authData.Expires))
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(500)
		return
	}

	refreshToken, _ := auth.MakeRefreshToken()
	args := database.CreateRefreshTokenParams{
		Token:     refreshToken,
		UserID:    userInfo.ID,
		ExpiresAt: time.Now().Add(time.Hour * 1440),
	}
	err = cfg.queries.CreateRefreshToken(r.Context(), args)
	if err != nil {
		fmt.Println(refreshToken)
		fmt.Println(err)
		w.WriteHeader(500)
		return
	}

	respData := User{
		ID:           userInfo.ID,
		CreatedAt:    userInfo.CreatedAt.Time,
		UpdatedAt:    userInfo.UpdatedAt.Time,
		Email:        userInfo.Email,
		Token:        token,
		RefreshToken: refreshToken,
		IsChirpyRed:  userInfo.IsChirpyRed.Bool,
	}
	respJson, err := json.Marshal(respData)
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(500)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(respJson)
}

func (cfg *apiConfig) HandlerRefresh(w http.ResponseWriter, r *http.Request) {
	type response struct {
		Token string `json:"token"`
	}

	refreshToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	userToken, err := cfg.queries.GetUserFromRefreshToken(r.Context(), refreshToken)
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(500)
		return
	}
	if userToken.RevokedAt.Valid || userToken.ExpiresAt.Before(time.Now()) {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	newAccessToken, err := auth.MakeJWT(userToken.UserID, cfg.secretJWT, time.Hour)
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(500)
		return
	}

	respToken := response{
		Token: newAccessToken,
	}
	respData, err := json.Marshal(respToken)
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(500)
		return
	}

	w.Write(respData)
}

func (cfg *apiConfig) HandlerRevoke(w http.ResponseWriter, r *http.Request) {
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	refreshToken, err := cfg.queries.GetUserFromRefreshToken(r.Context(), token)
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(500)
		return
	}

	err = cfg.queries.RevokeRefreshToken(r.Context(), refreshToken.UserID)
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(500)
		return
	}

	w.WriteHeader(204)
}

func (cfg *apiConfig) HandlerUpdateUser(w http.ResponseWriter, r *http.Request) {
	type authorizationInfo struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	type User struct {
		ID        uuid.UUID `json:"id"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		Email     string    `json:"email"`
	}

	decoder := json.NewDecoder(r.Body)
	insertData := authorizationInfo{}
	err := decoder.Decode(&insertData)
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(500)
		return
	}
	
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	realUserID, err := auth.ValidateJWT(token, cfg.secretJWT)
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	hashedPassword, err := auth.HashPassword(insertData.Password)
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(500)
		return
	}

	args := database.UpdateEmailPasswordParams{
		Email: insertData.Email,
		HashedPassword: hashedPassword,
		ID: realUserID,
	}
	user, err := cfg.queries.UpdateEmailPassword(r.Context(), args)
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(500)
		return
	}

	returnUser := User{
		ID: args.ID,
		CreatedAt: user.CreatedAt.Time,
		UpdatedAt: user.UpdatedAt.Time,
		Email: args.Email,
	}
	respData, err := json.Marshal(returnUser)
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(500)
		return
	}

	w.Write(respData)
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
		w.WriteHeader(403)
		return
	}

	err = cfg.queries.DeleteChirp(r.Context(), chirpID)
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(500)
		return
	}

	w.WriteHeader(204)
}

func (cfg *apiConfig) HandlerPolka(w http.ResponseWriter, r *http.Request) {
	type data struct {
		UserID uuid.UUID `json:"user_id"`
	}

	type req struct {
		Event string `json:"event"`
		Data  data   `json:"data"`
	}

	decoder := json.NewDecoder(r.Body)
	insertData := req{}
	err := decoder.Decode(&insertData)
	if err != nil{
		fmt.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if insertData.Event != "user.upgraded" {
		w.WriteHeader(204)
		return
	}

	err = cfg.queries.SetChirpyRed(r.Context(), insertData.Data.UserID)
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusNotFound)
		return
	}

	w.WriteHeader(204)
}