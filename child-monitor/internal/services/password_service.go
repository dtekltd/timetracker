package services

import (
	"errors"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

const bcryptCost = 12

// HasPassword returns true if a password hash is stored.
func HasPassword() (bool, error) {
	hash, err := GetSetting("password_hash")
	if err != nil {
		return false, err
	}
	return hash != "", nil
}

// SetPassword hashes and stores the password. Fails if one already exists.
func SetPassword(password string) error {
	if len(password) < 6 {
		return errors.New("password must be at least 6 characters")
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcryptCost)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}
	return SetSetting("password_hash", string(hash))
}

// VerifyPassword returns true if the provided password matches the stored hash.
func VerifyPassword(password string) (bool, error) {
	hash, err := GetSetting("password_hash")
	if err != nil {
		return false, err
	}
	if hash == "" {
		return false, errors.New("no password set")
	}
	err = bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
		return false, nil
	}
	return err == nil, err
}

// ChangePassword verifies the old password then replaces the hash.
func ChangePassword(oldPassword, newPassword string) error {
	ok, err := VerifyPassword(oldPassword)
	if err != nil {
		return err
	}
	if !ok {
		return errors.New("incorrect current password")
	}
	return SetPassword(newPassword)
}
