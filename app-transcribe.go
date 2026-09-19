// Copyright (c) 2024
// Licensed under the MIT License. See LICENSE file in the root directory.

package main

import (
	"fmt"
	"myscript/internal/repository"
	"myscript/internal/stt"
	"myscript/internal/transcribe/groq"
	"myscript/internal/transcribe/languages"
	"myscript/internal/transcribe/openai"
	witai "myscript/internal/transcribe/wait.ai"
)

// GetLanguages is what the configured source can transcribe; for a local
// model that is whatever the selected model declares.
func (a *App) GetLanguages() []languages.Language {
	config := a.GetConfig()

	switch config.TranscriberSource {
	case "witai":
		return witai.GetSupportedLanguages()
	case "local":
		if config.SpeechModelID != nil {
			if model, ok := stt.FindModel(*config.SpeechModelID); ok {
				return languages.Named(model.Languages)
			}
		}
		return nil
	default:
		return languages.Whisper
	}
}

// IsWitAIAvailable reports whether this build embeds Wit.ai keys; without them
// the option is hidden rather than offered and failing at runtime.
func (a *App) IsWitAIAvailable() bool {
	return witai.Available()
}

func (a *App) remoteTranscriber(source string) (stt.Transcriber, error) {
	secrets := a.secrets()

	switch source {
	case "witai":
		return func(wav []byte, language string) (string, error) {
			token, ok := witai.Token(language)
			if !ok {
				return "", fmt.Errorf("this build has no Wit.ai key for %s", language)
			}
			return witai.WitAITranscribeFromBuffer(wav, token)
		}, nil

	case "openai":
		apiKey := secrets.Get(repository.SecretSpeechOpenAIAPIKey)
		if apiKey == "" {
			return nil, fmt.Errorf("no OpenAI API key found")
		}
		return func(wav []byte, language string) (string, error) {
			return openai.TranscribeFromBuffer(wav, language, apiKey)
		}, nil

	case "groq":
		apiKey := secrets.Get(repository.SecretSpeechGroqAPIKey)
		if apiKey == "" {
			return nil, fmt.Errorf("no Groq API key found")
		}
		return func(wav []byte, language string) (string, error) {
			return groq.TranscribeFromBuffer(wav, language, apiKey)
		}, nil
	}

	return nil, fmt.Errorf("invalid transcriber source: %s", source)
}
