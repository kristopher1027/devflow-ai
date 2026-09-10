package auth

import "testing"

func TestHashPassword(t *testing.T) {
	password := "correct horse battery staple"

	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}

	if hash == "" {
		t.Fatal("expected password hash to be non-empty")
	}

	if hash == password {
		t.Fatal("password must not be stored as plaintext")
	}
}

func TestHashPasswordUsesDifferentSalts(t *testing.T) {
	password := "correct horse battery staple"

	hash1, err := HashPassword(password)
	if err != nil {
		t.Fatalf("hash password 1: %v", err)
	}

	hash2, err := HashPassword(password)
	if err != nil {
		t.Fatalf("hash password 2: %v", err)
	}

	if hash1 == hash2 {
		t.Fatal("expected different hashes for the same password")
	}
}

func TestVerifyPassword(t *testing.T) {
	password := "correct horse battery staple"

	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}

	if err := VerifyPassword(password, hash); err != nil {
		t.Fatalf("verify password: %v", err)
	}
}

func TestVerifyPasswordWrongPassword(t *testing.T) {
	hash, err := HashPassword("correct password")
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}

	if err := VerifyPassword("wrong password", hash); err == nil {
		t.Fatal("expected wrong password to fail")
	}
}

func TestHashPasswordEmpty(t *testing.T) {
	_, err := HashPassword("")

	if err == nil {
		t.Fatal("expected empty password to fail")
	}
}

func TestVerifyPasswordInvalidHash(t *testing.T) {
	err := VerifyPassword(
		"password",
		"not-a-valid-argon2id-hash",
	)

	if err == nil {
		t.Fatal("expected invalid hash to fail")
	}
}
