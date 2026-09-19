// Copyright (c) 2024
// Licensed under the MIT License. See LICENSE file in the root directory.

package remote

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"myscript/internal/transcribe/languages"
	"net/http"
	"strings"
	"sync"
)

// Gemini takes the audio inline, so the request carries the whole recording
// base64-encoded and the service caps the request rather than the file.
const geminiMaxRequestBytes = 18 * 1024 * 1024

const geminiPrompt = "Transcribe this recording exactly. Reply with the transcript only, " +
	"with no commentary, labels or timestamps."

// thinkingRefused remembers endpoint/model pairs that rejected the field, so
// the wasted round trip happens once per launch.
var thinkingRefused sync.Map

type geminiPart struct {
	Text       string            `json:"text,omitempty"`
	InlineData *geminiInlineData `json:"inline_data,omitempty"`
}

type geminiInlineData struct {
	MimeType string `json:"mime_type"`
	Data     string `json:"data"`
}

type geminiThinking struct {
	ThinkingBudget int `json:"thinkingBudget"`
}

type geminiGeneration struct {
	Temperature    float64         `json:"temperature"`
	ThinkingConfig *geminiThinking `json:"thinkingConfig,omitempty"`
}

type geminiRequest struct {
	Contents []struct {
		Parts []geminiPart `json:"parts"`
	} `json:"contents"`
	GenerationConfig geminiGeneration `json:"generationConfig"`
}

func geminiTranscribe(ctx context.Context, settings Settings, wav []byte, language string) (string, error) {
	encoded := base64.StdEncoding.EncodeToString(wav)
	if len(encoded) > geminiMaxRequestBytes {
		return "", fmt.Errorf("the recording is too long for Gemini")
	}

	request := geminiRequest{}
	request.Contents = append(request.Contents, struct {
		Parts []geminiPart `json:"parts"`
	}{
		Parts: []geminiPart{
			{Text: geminiPromptFor(language)},
			{InlineData: &geminiInlineData{MimeType: "audio/wav", Data: encoded}},
		},
	})

	key := settings.BaseURL + "|" + strings.ToLower(settings.Model)
	_, refused := thinkingRefused.Load(key)

	text, err := sendGemini(ctx, settings, request, !refused)
	if refused || !isBadRequest(err) {
		return text, err
	}

	// Matched on status, never message text. Remembered only once dropping the
	// field is confirmed to be what fixed it.
	text, retryErr := sendGemini(ctx, settings, request, false)
	if retryErr != nil {
		return "", retryErr
	}
	thinkingRefused.Store(key, true)
	return text, nil
}

func geminiPromptFor(language string) string {
	if language == "" {
		return geminiPrompt
	}
	return fmt.Sprintf("%s The speech is in %s; write the transcript in that language.",
		geminiPrompt, languages.Name(language))
}

func sendGemini(ctx context.Context, settings Settings, request geminiRequest, quiet bool) (string, error) {
	request.GenerationConfig = geminiGeneration{Temperature: 0}
	if quiet {
		request.GenerationConfig.ThinkingConfig = &geminiThinking{ThinkingBudget: 0}
	}

	payload, err := json.Marshal(request)
	if err != nil {
		return "", err
	}

	url := fmt.Sprintf("%s/v1beta/models/%s:generateContent", settings.BaseURL, settings.Model)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(payload))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-goog-api-key", settings.APIKey)

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("could not reach Gemini: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", &StatusError{
			Status:  resp.StatusCode,
			Message: fmt.Sprintf("Gemini returned %s: %s", resp.Status, errorMessage(body, resp.Status)),
		}
	}

	return parseGeminiResponse(body)
}

func parseGeminiResponse(body []byte) (string, error) {
	var parsed struct {
		Candidates []struct {
			Content struct {
				Parts []geminiPart `json:"parts"`
			} `json:"content"`
		} `json:"candidates"`
		PromptFeedback struct {
			BlockReason string `json:"blockReason"`
		} `json:"promptFeedback"`
	}

	if err := json.Unmarshal(body, &parsed); err != nil {
		return "", fmt.Errorf("unreadable response from Gemini: %w", err)
	}
	if reason := parsed.PromptFeedback.BlockReason; reason != "" {
		return "", fmt.Errorf("Gemini refused the recording (%s)", reason)
	}
	if len(parsed.Candidates) == 0 {
		return "", nil
	}

	var text strings.Builder
	for _, part := range parsed.Candidates[0].Content.Parts {
		text.WriteString(part.Text)
	}
	return strings.TrimSpace(text.String()), nil
}

func isBadRequest(err error) bool {
	var status *StatusError
	return errors.As(err, &status) && status.Status == http.StatusBadRequest
}
