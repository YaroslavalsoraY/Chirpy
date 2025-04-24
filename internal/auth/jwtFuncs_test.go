package auth

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestJWT(t *testing.T) {
	userID := uuid.New()
	token, err := MakeJWT(userID, "Salam", 1 * time.Minute)
	if err != nil {
		t.Errorf("ERROR: %v", err)
	}

	validID, err := ValidateJWT(token, "Salam")
	if err != nil {
		t.Errorf("ERROR: %v", err)
	}

	if userID != validID {
		t.Errorf("invalid id was returned")
	}
}

func TestJWT2(t *testing.T) {
	userID := uuid.New()
	token, err := MakeJWT(userID, "Salam", 1 * time.Minute)
	if err != nil {
		t.Errorf("ERROR: %v", err)
	}

	_, err = ValidateJWT(token, "1234")
	if err == nil {
		t.Errorf("invalid password was passed")
	}
}

func TestJWT3(t *testing.T) {
	userID := uuid.New()
	token, err := MakeJWT(userID, "Salam", 1 * time.Second)
	if err != nil {
		t.Errorf("ERROR: %v", err)
	}

	time.Sleep(1 * time.Second)

	_, err = ValidateJWT(token, "Salam")
	if err == nil {
		t.Errorf("token was passed when time is up")
	}
}