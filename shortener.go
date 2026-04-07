package main

import (
	"crypto/rand"
	"encoding/base64"
)

const codeLength = 6

// GenerateCode creates a random URL-safe short code like "aB3xZ9".
//
// We use crypto/rand (not math/rand) because math/rand is predictable —
// a user could guess future codes. crypto/rand reads from the OS's
// cryptographically secure random number generator.

func GenerateCode() (string, error) {
	bytes := make([]byte, codeLength)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	code := base64.RawURLEncoding.EncodeToString(bytes)
	return code[:codeLength], nil
}

// GenerateUniqueCode retries until it finds a code not already in the store.
func GenerateUniqueCode(store *Store) (string, error) {
	for {
		code, err := GenerateCode()
		if err != nil {
			return "", err
		}
		if !store.Exists(code) {
			return code, nil
		}
	}
}
