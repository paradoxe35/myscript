// Copyright (c) 2024
// Licensed under the MIT License. See LICENSE file in the root directory.

package main

import (
	"fmt"
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

func (a *App) remoteTranscriber(source string) (stt.Transcriber, error) {
	config := a.GetConfig()

	switch source {
	case "witai":
		return func(wav []byte, language string) (string, error) {
			apiKey := witai.GetAPIKey(language)
			if apiKey == nil {
				return "", fmt.Errorf("no API key found for language %s", language)
			}
			return witai.WitAITranscribeFromBuffer(wav, apiKey.Key)
		}, nil

	case "openai":
		if config.OpenAIApiKey == nil || *config.OpenAIApiKey == "" {
			return nil, fmt.Errorf("no OpenAI API key found")
		}
		apiKey := *config.OpenAIApiKey
		return func(wav []byte, language string) (string, error) {
			return openai.TranscribeFromBuffer(wav, language, apiKey)
		}, nil

	case "groq":
		if config.GroqApiKey == nil || *config.GroqApiKey == "" {
			return nil, fmt.Errorf("no Groq API key found")
		}
		apiKey := *config.GroqApiKey
		return func(wav []byte, language string) (string, error) {
			return groq.TranscribeFromBuffer(wav, language, apiKey)
		}, nil
	}

	return nil, fmt.Errorf("invalid transcriber source: %s", source)
}
