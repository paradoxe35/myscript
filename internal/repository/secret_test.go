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
	secrets.Set(SecretSpeechGroqAPIKey, "gsk_key")

	if err := secrets.Set(SecretSpeechGroqAPIKey, ""); err != nil {
		t.Fatal(err)
	}
	if secrets.Has(SecretSpeechGroqAPIKey) {
		t.Error("an empty value should clear the secret")
	}
	if got := secrets.Get(SecretSpeechGroqAPIKey); got != "" {
		t.Errorf("got %q", got)
	}
}

func TestMissingSecretsReadAsEmpty(t *testing.T) {
	secrets := NewSecretRepository(newDB(t, &Secret{}))

	if got := secrets.Get("nothing.here"); got != "" {
		t.Errorf("got %q", got)
	}
}

func TestAdoptLegacyKeysMovesThemOutOfTheSyncedConfig(t *testing.T) {
	mainDB, unsynced := newStores(t)

	notion, openAI, groq := "notion-key", "sk-openai", "gsk_groq"
	configs := NewConfigRepository(mainDB)
	configs.SaveConfig(&Config{NotionApiKey: &notion, OpenAIApiKey: &openAI, GroqApiKey: &groq})

	AdoptLegacyKeys(mainDB, unsynced)

	secrets := NewSecretRepository(unsynced)
	for name, want := range map[string]string{
		SecretNotionAPIKey:         notion,
		SecretSpeechOpenAIAPIKey:   openAI,
		SecretSpeechGroqAPIKey:     groq,
		AIProviderSecret("openai"): openAI,
	} {
		if got := secrets.Get(name); got != want {
			t.Errorf("secret %s = %q, want %q", name, got, want)
		}
	}

	config := configs.GetConfig()
	if config.NotionApiKey != nil || config.OpenAIApiKey != nil || config.GroqApiKey != nil {
		t.Error("the synced config should no longer carry credentials")
	}
}

func TestAdoptLegacyKeysKeepsWhatIsAlreadyStored(t *testing.T) {
	mainDB, unsynced := newStores(t)

	secrets := NewSecretRepository(unsynced)
	secrets.Set(SecretNotionAPIKey, "current-key")

	stale := "stale-key"
	NewConfigRepository(mainDB).SaveConfig(&Config{NotionApiKey: &stale})

	AdoptLegacyKeys(mainDB, unsynced)

	if got := secrets.Get(SecretNotionAPIKey); got != "current-key" {
		t.Errorf("got %q, a restored backup must not overwrite the local key", got)
	}
}

func TestAdoptLegacyKeysDoesNothingWithoutThem(t *testing.T) {
	mainDB, unsynced := newStores(t)
	NewConfigRepository(mainDB).SaveConfig(&Config{TranscriberSource: "local"})

	AdoptLegacyKeys(mainDB, unsynced)

	var count int64
	unsynced.Model(&Secret{}).Count(&count)
	if count != 0 {
		t.Errorf("stored %d secrets, want none", count)
	}
}
