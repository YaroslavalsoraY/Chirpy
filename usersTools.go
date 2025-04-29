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

type info struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	expires  int
}

type User struct {
	ID           uuid.UUID `json:"id"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	Email        string    `json:"email"`
	Token        string    `json:"token"`
	RefreshToken string    `json:"refresh_token"`
	IsChirpyRed  bool      `json:"is_chirpy_red"`
}

func (cfg *apiConfig) HandlerAddUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	decoder := json.NewDecoder(r.Body)
	user := info{}
	err := decoder.Decode(&user)
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	hash, err := auth.HashPassword(user.Password)
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	args := database.CreateUserParams{
		Email:          user.Email,
		HashedPassword: hash,
	}

	newUser, err := cfg.queries.CreateUser(r.Context(), args)
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	respUser := User{
		ID:          newUser.ID,
		CreatedAt:   newUser.CreatedAt.Time,
		UpdatedAt:   newUser.UpdatedAt.Time,
		Email:       newUser.Email,
		IsChirpyRed: false,
	}

	resp, err := json.Marshal(respUser)
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write(resp)
	return
}

func (cfg *apiConfig) HandlerLogin(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	authData := info{
		expires: 3600,
	}
	err := decoder.Decode(&authData)
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
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

	token, err := auth.MakeJWT(userInfo.ID, cfg.secretJWT, time.Second*time.Duration(authData.expires))
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
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
		w.WriteHeader(http.StatusInternalServerError)
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
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(respJson)
}

func (cfg *apiConfig) HandlerUpdateUser(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	insertData := info{}
	err := decoder.Decode(&insertData)
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
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
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	args := database.UpdateEmailPasswordParams{
		Email:          insertData.Email,
		HashedPassword: hashedPassword,
		ID:             realUserID,
	}
	user, err := cfg.queries.UpdateEmailPassword(r.Context(), args)
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	returnUser := User{
		ID:        args.ID,
		CreatedAt: user.CreatedAt.Time,
		UpdatedAt: user.UpdatedAt.Time,
		Email:     args.Email,
	}
	respData, err := json.Marshal(returnUser)
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Write(respData)
}
