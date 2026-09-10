package auth

import (
	"crypto/sha256"
	"encoding/hex"
)

func HashSessionToken(token string) string {
	hash := sha256.Sum256([]byte(token))

	return hex.EncodeToString(hash[:])
}
