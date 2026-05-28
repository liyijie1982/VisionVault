package service

import (
	"errors"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

var ErrPasswordRequired = errors.New("password is required")
var ErrPasswordTooShort = errors.New("password must be at least 8 characters")
var ErrCurrentPasswordIncorrect = errors.New("current password is incorrect")

func normalizePassword(password string) (string, error) {
	password = strings.TrimSpace(password)
	if password == "" {
		return "", ErrPasswordRequired
	}
	if len(password) < 8 {
		return "", ErrPasswordTooShort
	}
	return password, nil
}

func hashPassword(password string) (string, error) {
	normalized, err := normalizePassword(password)
	if err != nil {
		return "", err
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(normalized), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashed), nil
}

func comparePassword(hash, password string) error {
	if strings.TrimSpace(hash) == "" {
		return ErrInvalidCredentials
	}
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(strings.TrimSpace(password)))
}
