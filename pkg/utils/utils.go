package utils

import (
	"crypto/rand"
	"encoding/base32"

	"golang.org/x/crypto/bcrypt"
)

func GenerateUid(length int) (string, error) {
	token_bytes := make([]byte, length)
	_, err := rand.Read(token_bytes)
	if err != nil {
		return "", err
	}
	token := base32.StdEncoding.EncodeToString(token_bytes)
	return token, nil
}

func GeneratePassword(password string) ([]byte, error) {
	return bcrypt.GenerateFromPassword([]byte(password), 12)
}
