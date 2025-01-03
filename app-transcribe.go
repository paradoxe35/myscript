package main

import (
	"fmt"
	"log"
	"myscript/internal/transcribe/structs"
	witai "myscript/internal/transcribe/wait.ai"
	"myscript/internal/transcribe/whisper/openai"
)

// --- WitAI ---

func (a *App) GetWitAILanguages() []structs.Language {
	return witai.GetSupportedLanguages()
}

// --- Languages ---

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

func (a *App) OpenAPITranscribe(buffer []byte, language string) (string, error) {
	config := a.GetConfig()

	if config.OpenAIApiKey == nil || *config.OpenAIApiKey == "" {
		return "", fmt.Errorf("no OpenAI API key found")
	}

	text, err := openai.TranscribeFromBuffer(buffer, language, *config.OpenAIApiKey)
	if err != nil {
		return "", err
	}

	return text, nil
}

func (a *App) LocalTranscribe(buffer []byte, language string) (string, error) {
	return a.lwt.Transcribe(buffer, language)
}

func (a *App) Transcribe(buffer []byte, language string) (string, error) {
	config := a.GetConfig()

	log.Printf("Transcribing with language: %s, source: %s", language, config.TranscriberSource)

	switch config.TranscriberSource {
	case "witai":
		return a.WitTranscribe(buffer, language)
	case "openai":
		return a.OpenAPITranscribe(buffer, language)
	case "local":
		return a.LocalTranscribe(buffer, language)
	}

	return "", fmt.Errorf("invalid transcriber source: %s", config.TranscriberSource)
}
