// Copyright (c) 2024
// Licensed under the MIT License. See LICENSE file in the root directory.

package main

import (
	"fmt"
	"log/slog"
	"myscript/internal/transcribe/groq"
	"myscript/internal/transcribe/structs"
	witai "myscript/internal/transcribe/wait.ai"
	"myscript/internal/transcribe/whisper/openai"
)

// --- Languages ---

func (a *App) GetWitAILanguages() []structs.Language {
	return witai.GetSupportedLanguages()
}

func (a *App) GetLanguages() []structs.Language {
	config := a.GetConfig()

	switch config.TranscriberSource {
	case "witai":
		return a.GetWitAILanguages()
	default:
		return a.GetWhisperLanguages()
	}
}

// --- Transcribe ---

func (a *App) initLocalWhisperTranscriber(language string) error {
	config := a.GetConfig()

	if config.TranscriberSource != "local" {
		return nil
	}

	var configuredModel string
	if config.LocalWhisperModel == nil {
		configuredModel = a.GetBestLocalWhisperModel()
	} else {
		configuredModel = *config.LocalWhisperModel
	}

	return a.lwt.LoadModel(configuredModel, language)
}

func (a *App) WitTranscribe(buffer []byte, language string) (string, error) {
	apiKey := witai.GetAPIKey(language)
	if apiKey == nil {
		return "", fmt.Errorf("no API key found for language %s", language)
	}

	return witai.WitAITranscribeFromBuffer(buffer, apiKey.Key)
}

func (a *App) OpenAITranscribe(buffer []byte, language string) (string, error) {
	config := a.GetConfig()
	if *config.OpenAIApiKey == "" {
		return "", fmt.Errorf("no OpenAI API key found")
	}

	if text, err := openai.TranscribeFromBuffer(buffer, language, *config.OpenAIApiKey); err != nil {
		return "", err
	} else {
		return text, nil
	}
}

func (a *App) GroqTranscribe(buffer []byte, language string) (string, error) {
	config := a.GetConfig()
	if *config.GroqApiKey == "" {
		return "", fmt.Errorf("no Groq API key found")
	}

	if text, err := groq.TranscribeFromBuffer(buffer, language, *config.GroqApiKey); err != nil {
		return "", err
	} else {
		return text, nil
	}
}

func (a *App) LocalTranscribe(buffer []byte, language string) (string, error) {
	return a.lwt.Transcribe(buffer, language)
}

func (a *App) Transcribe(buffer []byte, language string) (string, error) {
	config := a.GetConfig()

	slog.Debug("Transcribing with language", "language", language, "source", config.TranscriberSource)

	switch config.TranscriberSource {
	case "witai":
		return a.WitTranscribe(buffer, language)
	case "openai":
		return a.OpenAITranscribe(buffer, language)
	case "local":
		return a.LocalTranscribe(buffer, language)
	case "groq":
		return a.GroqTranscribe(buffer, language)
	}

	return "", fmt.Errorf("invalid transcriber source: %s", config.TranscriberSource)
}
