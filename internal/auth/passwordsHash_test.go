package auth

import(
	"testing"
)

func TestCheckPasswordHash(t *testing.T){
	hash, err := HashPassword("bodybuilding123")
	if err != nil {
		t.Errorf("ERROR: %v", err)
	}

	err = CheckPasswordHash(hash, "bodybuilding123")
	if err != nil {
		t.Errorf("ERROR: %v", err)
	}
}

func TestCheckPasswordHash2(t *testing.T){
	hash, err := HashPassword("SalamPopolam")
	if err != nil {
		t.Errorf("ERROR: %v", err)
	}

	err = CheckPasswordHash(hash, "bodybuilding123")
	if err == nil {
		t.Errorf("ERROR: wrong password passed")
	}
}