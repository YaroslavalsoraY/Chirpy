package auth

import (
	"net/http"
	"testing"
)

func TestGetBearerToken(t *testing.T) {
	header := http.Header{"Authorization": []string{"Bearer TOKEN_STRING"}}

	token, err := GetBearerToken(header)
	if err != nil {
		t.Errorf("ERROR: %v", err)
	}

	if token != "TOKEN_STRING" {
		t.Errorf("invalid result")
	}
}