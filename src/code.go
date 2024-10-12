package main

import (
	"crypto/rand"
	"encoding/base32"
)

func generateSecureCode() (string, error) {
	bytes := make([]byte, 5)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}
	// Remove 0, O, 1, I to remove ambiguity
	code := base32.NewEncoding("ABCDEFGHJKLMNPQRSTUVWXYZ23456789").EncodeToString(bytes)
	return code, nil
}
