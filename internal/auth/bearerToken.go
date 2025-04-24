package auth

import (
	"errors"
	"net/http"
	"strings"
)

func GetBearerToken(headers http.Header) (string, error){
	rawToken := headers.Clone().Get("Authorization")
	if rawToken == "" {
		return "", errors.New("Token not found")
	}

	resToken := strings.ReplaceAll(rawToken, "Bearer ", "")

	return resToken, nil
}