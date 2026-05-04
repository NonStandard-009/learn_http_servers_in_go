package auth

import (
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
