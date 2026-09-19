// Copyright (c) 2024
// Licensed under the MIT License. See LICENSE file in the root directory.

package remote

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strings"
	"time"
)

const (
	// OpenAI and Groq's free tier both cap uploads here.
	maxAudioBytes  = 25 * 1024 * 1024
	requestTimeout = 3 * time.Minute
)

var client = &http.Client{Timeout: requestTimeout}

// Transcribe sends a WAV recording to the configured service.
func Transcribe(ctx context.Context, settings Settings, wav []byte, language string) (string, error) {
	settings = settings.Resolved()

	if settings.Model == "" {
		return "", fmt.Errorf("no transcription model is configured")
	}
	if settings.BaseURL == "" {
		return "", fmt.Errorf("no endpoint is configured")
	}

	if protocolFor(settings.Preset) == ProtocolGemini {
		return geminiTranscribe(ctx, settings, wav, language)
	}
	return openAITranscribe(ctx, settings, wav, language)
}

func openAITranscribe(ctx context.Context, settings Settings, wav []byte, language string) (string, error) {
	if len(wav) > maxAudioBytes {
		return "", fmt.Errorf("the recording is too large for this service (limit is %d MB)", maxAudioBytes/(1024*1024))
	}

	body, contentType, err := multipartBody(settings, wav, language)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, settings.BaseURL+"/audio/transcriptions", body)
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", contentType)
	// A local runtime takes no credentials, and some reject an empty bearer.
	if settings.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+settings.APIKey)
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("could not reach the transcription service: %w", err)
	}
	defer resp.Body.Close()

	payload, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", &StatusError{Status: resp.StatusCode, Message: errorMessage(payload, resp.Status)}
	}

	var parsed struct {
		Text string `json:"text"`
	}
	if err := json.Unmarshal(payload, &parsed); err != nil {
		return "", fmt.Errorf("unreadable response from the transcription service: %w", err)
	}
	return parsed.Text, nil
}

func multipartBody(settings Settings, wav []byte, language string) (io.Reader, string, error) {
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	part, err := writer.CreateFormFile("file", "audio.wav")
	if err != nil {
		return nil, "", err
	}
	if _, err := part.Write(wav); err != nil {
		return nil, "", err
	}

	if err := writer.WriteField("model", settings.Model); err != nil {
		return nil, "", err
	}
	if language != "" {
		if err := writer.WriteField(languageField(settings.Model), language); err != nil {
			return nil, "", err
		}
	}

	if err := writer.Close(); err != nil {
		return nil, "", err
	}
	return &buf, writer.FormDataContentType(), nil
}

// The gpt- transcribe models take a repeated languages[]; whisper-1, every Groq
// model and anything unrecognised take language.
func languageField(model string) string {
	if strings.HasPrefix(strings.ToLower(model), "gpt-") {
		return "languages[]"
	}
	return "language"
}

// StatusError carries the status so a caller can react to a 400 without
// matching on wording every service spells differently.
type StatusError struct {
	Status  int
	Message string
}

func (e *StatusError) Error() string { return e.Message }

// errorMessage reads both shapes in use: {"error":{"message":...}} and {"error":"..."}.
func errorMessage(body []byte, status string) string {
	var wrapped struct {
		Error struct {
			Message string `json:"message"`
		} `json:"error"`
	}
	if json.Unmarshal(body, &wrapped) == nil && wrapped.Error.Message != "" {
		return wrapped.Error.Message
	}

	var flat struct {
		Error string `json:"error"`
	}
	if json.Unmarshal(body, &flat) == nil && flat.Error != "" {
		return flat.Error
	}

	if trimmed := strings.TrimSpace(string(body)); trimmed != "" {
		return trimmed
	}
	return status
}
