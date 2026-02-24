package utils

import (
	"crypto/sha256"
	"encoding/base64"
)

func GenerateShortURL(originalURL string) (string, error) {
	hash := sha256.Sum256([]byte(originalURL))
	encoded := base64.URLEncoding.EncodeToString(hash[:])
	if len(encoded) > 10 {
		encoded = encoded[len(encoded)-10:]
	}
	return encoded, nil
}
