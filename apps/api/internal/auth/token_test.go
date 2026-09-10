package auth

import (
	"testing"
)

func TestGenerateSessionToken(t *testing.T) {
	token1, err := GenerateSessionToken()
	if err != nil {
		t.Fatalf("generate token 1: %v", err)
	}

	token2, err := GenerateSessionToken()
	if err != nil {
		t.Fatalf("generate token 2: %v", err)
	}

	if token1 == "" {
		t.Fatal("expected token 1 to be non-empty")
	}

	if token2 == "" {
		t.Fatal("expected token 2 to be non-empty")
	}

	if token1 == token2 {
		t.Fatal("expected generated tokens to be different")
	}

	if len(token1) != sessionTokenBytes*2 {
		t.Fatalf(
			"expected token length %d, got %d",
			sessionTokenBytes*2,
			len(token1),
		)
	}
}
