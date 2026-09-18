// Copyright (c) 2024
// Licensed under the MIT License. See LICENSE file in the root directory.

package secrets

import (
	"errors"
	"testing"
)

func TestRoundTrip(t *testing.T) {
	encrypted, err := Encrypt("sk-secret")
	if err != nil {
		t.Fatal(err)
	}
	if encrypted == "sk-secret" {
		t.Fatal("the value was stored in the clear")
	}

	decrypted, err := Decrypt(encrypted)
	if err != nil {
		t.Fatal(err)
	}
	if decrypted != "sk-secret" {
		t.Errorf("got %q", decrypted)
	}
}

func TestEncryptUsesAFreshNonce(t *testing.T) {
	first, _ := Encrypt("same")
	second, _ := Encrypt("same")

	if first == second {
		t.Error("equal plaintexts must not produce equal ciphertexts")
	}
}

func TestEmptyValuesPassThrough(t *testing.T) {
	if encrypted, _ := Encrypt(""); encrypted != "" {
		t.Errorf("Encrypt(\"\") = %q", encrypted)
	}
	if decrypted, _ := Decrypt(""); decrypted != "" {
		t.Errorf("Decrypt(\"\") = %q", decrypted)
	}
}

func TestDecryptRejectsTamperedValues(t *testing.T) {
	encrypted, _ := Encrypt("sk-secret")
	tampered := "A" + encrypted[1:]

	if _, err := Decrypt(tampered); !errors.Is(err, ErrUnreadable) {
		t.Fatalf("got %v, want ErrUnreadable", err)
	}
	if _, err := Decrypt("not base64!"); !errors.Is(err, ErrUnreadable) {
		t.Fatalf("got %v, want ErrUnreadable", err)
	}
}
