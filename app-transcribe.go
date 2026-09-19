// Copyright (c) 2024
// Licensed under the MIT License. See LICENSE file in the root directory.

package main

import (
	"context"
	"fmt"
	"myscript/internal/repository"
	"myscript/internal/stt"
	"myscript/internal/transcribe/languages"
	"myscript/internal/transcribe/remote"
	witai "myscript/internal/transcribe/wait.ai"
)

// SpeechService is a hosted transcription service as the settings screen lists it.
type SpeechService struct {
	ID      string
	Name    string
	BaseURL string
	Models  []string
	KeyHint string
	Custom  bool
}

func (a *App) GetSpeechServices() []SpeechService {
	services := make([]SpeechService, 0, len(remote.Presets))

	for _, preset := range remote.Presets {
		services = append(services, SpeechService{
			ID:      preset.ID,
			Name:    preset.Name,
			BaseURL: preset.BaseURL,
			Models:  preset.Models,
			KeyHint: preset.KeyHint,
			Custom:  preset.Custom(),
		})
	}
	return services
}

// IsWitAIAvailable reports whether this build embeds Wit.ai keys; without them
// the option is hidden rather than offered and failing at runtime.
func (a *App) IsWitAIAvailable() bool {
	return witai.Available()
}

func (a *App) GetSpeechServiceAPIKey(preset string) string {
	return a.secrets().Get(repository.SpeechServiceSecret(preset))
}

func (a *App) SaveSpeechServiceAPIKey(preset, apiKey string) error {
	return a.secrets().Set(repository.SpeechServiceSecret(preset), apiKey)
}

// GetLanguages is what the configured source can transcribe. An empty list
// means the codes are not known, so the language should be typed.
func (a *App) GetLanguages() []languages.Language {
	config := a.GetConfig()

	switch config.TranscriberSource {
	case "witai":
		return witai.GetSupportedLanguages()

	case "remote":
		return remote.LanguagesFor(config.RemoteProvider, config.RemoteModel)

	default:
		if config.SpeechModelID == nil {
			return nil
		}
		model, ok := stt.FindModel(*config.SpeechModelID)
		if !ok {
			return nil
		}
		return languages.Named(model.Languages)
	}
}

func (a *App) remoteTranscriber(source string) (stt.Transcriber, error) {
	switch source {
	case "witai":
		return func(wav []byte, language string) (string, error) {
			token, ok := witai.Token(language)
			if !ok {
				return "", fmt.Errorf("this build has no Wit.ai key for %s", language)
			}
			return witai.WitAITranscribeFromBuffer(wav, token)
		}, nil

	case "remote":
		settings := a.speechServiceSettings()
		if !settings.Ready() {
			return nil, fmt.Errorf("the transcription service needs an endpoint and a model")
		}
		return func(wav []byte, language string) (string, error) {
			return remote.Transcribe(context.Background(), settings, wav, language)
		}, nil
	}

	return nil, fmt.Errorf("invalid transcriber source: %s", source)
}

func (a *App) speechServiceSettings() remote.Settings {
	config := a.GetConfig()

	return remote.Settings{
		Preset:  config.RemoteProvider,
		BaseURL: config.RemoteBaseURL,
		Model:   config.RemoteModel,
		APIKey:  a.GetSpeechServiceAPIKey(config.RemoteProvider),
	}.Resolved()
}
