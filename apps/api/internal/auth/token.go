package auth

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
)

const sessionTokenBytes = 32

func GenerateSessionToken() (string, error) {
	token := make([]byte, sessionTokenBytes)

	if _, err := rand.Read(token); err != nil {
		return "", fmt.Errorf("generate session token: %w", err)
	}

	return hex.EncodeToString(token), nil
}