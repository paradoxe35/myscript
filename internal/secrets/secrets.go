// Copyright (c) 2024
// Licensed under the MIT License. See LICENSE file in the root directory.

// Package secrets encrypts credentials at rest with a machine-derived key, so
// a copied database file carries nothing usable.
package secrets

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"os"
	"runtime"
	"sync"
)

var (
	keyOnce sync.Once
	key     []byte
)

func machineKey() []byte {
	keyOnce.Do(func() {
		hostname, _ := os.Hostname()
		user := os.Getenv("USER")
		if user == "" {
			user = os.Getenv("USERNAME")
		}

		sum := sha256.Sum256([]byte(fmt.Sprintf("%s-%s-%s", hostname, user, runtime.GOOS)))
		key = sum[:]
	})
	return key
}

func newGCM() (cipher.AEAD, error) {
	block, err := aes.NewCipher(machineKey())
	if err != nil {
		return nil, err
	}
	return cipher.NewGCM(block)
}

func Encrypt(plaintext string) (string, error) {
	if plaintext == "" {
		return "", nil
	}

	gcm, err := newGCM()
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(gcm.Seal(nonce, nonce, []byte(plaintext), nil)), nil
}

var ErrUnreadable = errors.New("the stored secret cannot be read on this machine")

func Decrypt(encoded string) (string, error) {
	if encoded == "" {
		return "", nil
	}

	raw, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return "", ErrUnreadable
	}

	gcm, err := newGCM()
	if err != nil {
		return "", err
	}
	if len(raw) < gcm.NonceSize() {
		return "", ErrUnreadable
	}

	nonce, ciphertext := raw[:gcm.NonceSize()], raw[gcm.NonceSize():]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", ErrUnreadable
	}

	return string(plaintext), nil
}
