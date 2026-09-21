// Copyright (c) 2024
// Licensed under the MIT License. See LICENSE file in the root directory.

package ai

import (
	"encoding/json"
	"fmt"
	"strings"
)

// StatusCode lets callers react to a status without matching provider wording.
type APIError struct {
	StatusCode int
	Message    string
}

func (e *APIError) Error() string { return e.Message }

func apiError(status int, format string, args ...any) error {
	return &APIError{StatusCode: status, Message: fmt.Sprintf(format, args...)}
}

func parseAPIError(status int, body []byte, provider string) error {
	if message := messageFromBody(body); message != "" {
		return apiError(status, "%s API error (%d): %s", provider, status, message)
	}

	switch status {
	case 401:
		return apiError(status, "%s rejected the API key", provider)
	case 403:
		return apiError(status, "%s refused access: check the key's permissions", provider)
	case 404:
		return apiError(status, "%s has no such endpoint: check the base URL and the model name", provider)
	case 429:
		return apiError(status, "%s rate limit reached: try again in a moment", provider)
	case 500, 502, 503, 504:
		return apiError(status, "%s is unavailable right now (%d)", provider, status)
	}

	if preview := truncate(string(body), 100); preview != "" {
		return apiError(status, "%s request failed (%d): %s", provider, status, preview)
	}

	return apiError(status, "%s request failed with status %d", provider, status)
}

func messageFromBody(body []byte) string {
	var payload struct {
		Error *struct {
			Message string `json:"message"`
		} `json:"error"`
	}

	if json.Unmarshal(body, &payload) != nil || payload.Error == nil {
		return ""
	}

	return payload.Error.Message
}

func decodeError(err error, body []byte, provider string) error {
	switch {
	case len(body) == 0:
		return fmt.Errorf("%s returned an empty response: check the base URL", provider)
	case strings.Contains(string(body), "<!DOCTYPE"), strings.Contains(string(body), "<html"):
		return fmt.Errorf("%s returned a web page instead of JSON: check the base URL", provider)
	default:
		return fmt.Errorf("%s returned an unreadable response: %w", provider, err)
	}
}

func truncate(text string, max int) string {
	if len(text) <= max {
		return text
	}
	return text[:max] + "..."
}
