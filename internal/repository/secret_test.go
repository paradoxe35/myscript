// Copyright (c) 2024
// Licensed under the MIT License. See LICENSE file in the root directory.

package repository

import "testing"

func TestSecretRoundTrip(t *testing.T) {
	secrets := NewSecretRepository(newDB(t, &Secret{}))

	if err := secrets.Set(SecretNotionAPIKey, "secret-token"); err != nil {
		t.Fatal(err)
	}

	if got := secrets.Get(SecretNotionAPIKey); got != "secret-token" {
		t.Errorf("got %q", got)
	}
	if !secrets.Has(SecretNotionAPIKey) {
		t.Error("Has should report a stored secret")
	}
}

func TestSecretIsEncryptedOnDisk(t *testing.T) {
	db := newDB(t, &Secret{})
	secrets := NewSecretRepository(db)
	secrets.Set(SecretNotionAPIKey, "secret-token")

	var stored Secret
	db.First(&stored)

	if stored.Value == "secret-token" {
		t.Fatal("the secret was written in the clear")
	}
}

func TestSettingAnEmptyValueRemovesTheSecret(t *testing.T) {
	secrets := NewSecretRepository(newDB(t, &Secret{}))
	secrets.Set(SpeechServiceSecret("groq"), "gsk_key")

	if err := secrets.Set(SpeechServiceSecret("groq"), ""); err != nil {
		t.Fatal(err)
	}
	if secrets.Has(SpeechServiceSecret("groq")) {
		t.Error("an empty value should clear the secret")
	}
	if got := secrets.Get(SpeechServiceSecret("groq")); got != "" {
		t.Errorf("got %q", got)
	}
}

func TestMissingSecretsReadAsEmpty(t *testing.T) {
	secrets := NewSecretRepository(newDB(t, &Secret{}))

	if got := secrets.Get("nothing.here"); got != "" {
		t.Errorf("got %q", got)
	}
}

func TestRecordingTheSameProcessedFileTwiceKeepsOneRow(t *testing.T) {
	db := newDB(t, &ProcessedChange{})
	repo := NewProcessedChangeRepository(db)

	for range 3 {
		if err := repo.SaveProcessedChange("file-1"); err != nil {
			t.Fatal(err)
		}
	}

	var rows int64
	db.Model(&ProcessedChange{}).Count(&rows)
	if rows != 1 {
		t.Errorf("got %d rows for one file, a cycle would add one every time", rows)
	}
	if !repo.ChangeProcessed("file-1") {
		t.Error("the file should still be recognised as processed")
	}
}
