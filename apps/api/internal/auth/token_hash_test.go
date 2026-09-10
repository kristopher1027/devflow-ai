package auth

import "testing"

func TestHashSessionToken(t *testing.T) {
	token := "test-session-token"

	hash := HashSessionToken(token)

	if hash == "" {
		t.Fatal("expected hash to be non-empty")
	}

	if len(hash) != 64 {
		t.Fatalf(
			"expected hash length 64, got %d",
			len(hash),
		)
	}

	if hash != HashSessionToken(token) {
		t.Fatal("expected same token to produce same hash")
	}

	if hash == HashSessionToken("different-token") {
		t.Fatal("expected different tokens to produce different hashes")
	}
}
