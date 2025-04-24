package auth

import (
	"crypto/rand"
	"encoding/hex"
)

func MakeRefreshToken() (string, error) {
	refreshToken := make([]byte, 32)
	rand.Read(refreshToken)
	return hex.EncodeToString(refreshToken), nil
}