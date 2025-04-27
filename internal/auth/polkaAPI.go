package auth

import (
	"errors"
	"net/http"
	"strings"
)

func GetPolkaAPI(headers http.Header) (string, error){
	rawKey := headers.Clone().Get("Authorization")
	if rawKey == "" {
		return "", errors.New("Incorrect API key")
	}

	resKey := strings.ReplaceAll(rawKey, "ApiKey ", "")

	return resKey, nil
}