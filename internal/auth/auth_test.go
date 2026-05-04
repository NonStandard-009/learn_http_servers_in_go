package auth

import (
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestMakeValidateJWT(t *testing.T) {
	testID := uuid.New()
	secretA := "Yes. this is literally a secret lol."
	secretB := "No, this one is not a secret..."

	testTokenStr, err := MakeJWT(testID, secretA, time.Minute*5)
	if err != nil {
		t.Fatalf("Unexpected error when creating token: %v", err)
	}

	resultID, err := ValidateJWT(testTokenStr, secretA)
	if err != nil {
		t.Fatalf("Unexpected error while validating token: %v", err)
	}

	if resultID != testID {
		t.Errorf("ID's do not match. Got: %v, want: %v", resultID, testID)
	}

	_, err = ValidateJWT(testTokenStr, secretB)
	if err == nil {
		t.Error("Unexpected success when validating token with INCORRECT secret.")
	}

	testExpiredTokenStr, err := MakeJWT(testID, secretA, time.Second*(-1))
	if err != nil {
		t.Fatalf("Unexpected error when creating token: %v", err)
	}

	_, err = ValidateJWT(testExpiredTokenStr, secretA)
	if err == nil {
		t.Error("Unexpected success when validating expired token.")
	}
}

func TestGetBearerToken(t *testing.T) {
	header := http.Header{}
	header.Add("Authorization", "Bearer 123this321")
	expectedToken := "123this321"

	actualTokenA, err := GetBearerToken(header)
	if err != nil {
		t.Fatalf("Unexpected error while getting bearer token: %v", err)
	}

	if actualTokenA != expectedToken {
		t.Errorf("Incorrect token. Got: %s, want: %s", actualTokenA, expectedToken)
	}

	header.Add("Authorization", "Bearer     123this321   ")
	actualTokenB, err := GetBearerToken(header)
	if err != nil {
		t.Fatalf("Unexpected error while getting bearer token: %v", err)
	}

	if actualTokenB != expectedToken {
		t.Errorf("Incorrect token. Got: %s, want: %s", actualTokenB, expectedToken)
	}
}

func TestGetBearerToken_MissingHeader(t *testing.T) {
	header := http.Header{}
	_, err := GetBearerToken(header)
	if err == nil {
		t.Fatal("expected an error when Authorization header is missing")
	}
}
