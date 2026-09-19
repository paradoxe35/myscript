// Copyright (c) 2024
// Licensed under the MIT License. See LICENSE file in the root directory.

package repository

import (
	"log/slog"
	"myscript/internal/ai"
	"strings"

	"gorm.io/gorm"
)

// AdoptLegacyKeys moves credentials written by older builds out of the synced
// Config and into the local secret store, then clears them so they stop being
// uploaded. Safe to call repeatedly: a restored backup can bring them back.
func AdoptLegacyKeys(mainDB, unSyncedDB *gorm.DB) {
	configs := NewConfigRepository(mainDB)
	secrets := NewSecretRepository(unSyncedDB)

	config := configs.GetConfig()

	legacy := map[string]*string{
		SecretNotionAPIKey:              config.NotionApiKey,
		SecretSpeechOpenAIAPIKey:        config.OpenAIApiKey,
		SecretSpeechGroqAPIKey:          config.GroqApiKey,
		AIProviderSecret(ai.KindOpenAI): config.OpenAIApiKey,
	}

	moved := false
	for name, value := range legacy {
		if adopt(secrets, name, value) {
			moved = true
		}
	}

	if !moved {
		return
	}

	config.NotionApiKey = nil
	config.OpenAIApiKey = nil
	config.GroqApiKey = nil
	configs.SaveConfig(config)

	slog.Info("Moved API keys out of the synced configuration into the local secret store")
}

// AdoptHostedSpeech moves a build that named its transcription service directly
// onto the hosted-service settings that replaced them.
func AdoptHostedSpeech(mainDB, unSyncedDB *gorm.DB) {
	configs := NewConfigRepository(mainDB)
	secrets := NewSecretRepository(unSyncedDB)

	config := configs.GetConfig()

	preset, legacyKey := "", ""
	switch config.TranscriberSource {
	case "openai":
		preset, legacyKey = "openai", SecretSpeechOpenAIAPIKey
	case "groq":
		preset, legacyKey = "groq", SecretSpeechGroqAPIKey
	default:
		return
	}

	if key := secrets.Get(legacyKey); key != "" && !secrets.Has(SpeechServiceSecret(preset)) {
		if err := secrets.Set(SpeechServiceSecret(preset), key); err != nil {
			slog.Error("Could not move a transcription key", "service", preset, "error", err)
			return
		}
	}

	config.TranscriberSource = "remote"
	config.RemoteProvider = preset
	configs.SaveConfig(config)

	slog.Info("Moved the transcription service onto the hosted settings", "service", preset)
}

func adopt(secrets *SecretRepository, name string, legacy *string) bool {
	if legacy == nil || strings.TrimSpace(*legacy) == "" {
		return false
	}
	if secrets.Has(name) {
		return true
	}

	if err := secrets.Set(name, strings.TrimSpace(*legacy)); err != nil {
		slog.Error("Could not store a legacy API key", "name", name, "error", err)
		return false
	}
	return true
}
