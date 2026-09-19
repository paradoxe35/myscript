// Copyright (c) 2024
// Licensed under the MIT License. See LICENSE file in the root directory.

package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
	"time"
)

const modelListTimeout = 20 * time.Second

type ModelInfo struct {
	ID   string
	Name string
}

type catalogEndpoint struct {
	url     string
	headers map[string]string
}

func catalogEndpointFor(settings Settings) catalogEndpoint {
	settings = settings.Resolved()
	base := settings.BaseURL
	if base == "" {
		base = DefaultBaseURL(settings.Kind)
	}

	switch settings.Kind {
	case KindGemini:
		return catalogEndpoint{
			url:     base + "/v1beta/models?pageSize=1000",
			headers: map[string]string{"x-goog-api-key": settings.APIKey},
		}

	default:
		endpoint := catalogEndpoint{url: base + "/models", headers: map[string]string{}}
		if settings.APIKey != "" {
			endpoint.headers["Authorization"] = "Bearer " + settings.APIKey
		}
		return endpoint
	}
}

func ListModels(ctx context.Context, settings Settings) ([]ModelInfo, error) {
	endpoint := catalogEndpointFor(settings)

	ctx, cancel := context.WithTimeout(ctx, modelListTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	for name, value := range endpoint.headers {
		req.Header.Set(name, value)
	}

	resp, err := newHTTPClient().Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to reach the provider: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, maxErrorBody))
		return nil, parseAPIError(resp.StatusCode, body, settings.Name)
	}

	models, err := decodeModels(resp.Body)
	if err != nil {
		return nil, err
	}

	sort.Slice(models, func(i, j int) bool { return models[i].ID < models[j].ID })
	return models, nil
}

// OpenAI and Anthropic answer with "data", Gemini with "models" and a prefixed id.
func decodeModels(r io.Reader) ([]ModelInfo, error) {
	var payload struct {
		Data []struct {
			ID          string `json:"id"`
			Name        string `json:"name"`
			DisplayName string `json:"display_name"`
		} `json:"data"`
		Models []struct {
			Name        string   `json:"name"`
			DisplayName string   `json:"displayName"`
			Methods     []string `json:"supportedGenerationMethods"`
		} `json:"models"`
	}

	if err := json.NewDecoder(r).Decode(&payload); err != nil {
		return nil, fmt.Errorf("failed to read the model list: %w", err)
	}

	models := make([]ModelInfo, 0, len(payload.Data)+len(payload.Models))

	for _, entry := range payload.Data {
		if entry.ID == "" {
			continue
		}
		name := entry.DisplayName
		if name == "" {
			name = entry.Name
		}
		models = append(models, ModelInfo{ID: entry.ID, Name: name})
	}

	for _, entry := range payload.Models {
		if !supportsGeneration(entry.Methods) {
			continue
		}
		models = append(models, ModelInfo{
			ID:   strings.TrimPrefix(entry.Name, "models/"),
			Name: entry.DisplayName,
		})
	}

	return models, nil
}

func supportsGeneration(methods []string) bool {
	if len(methods) == 0 {
		return true
	}
	for _, method := range methods {
		if method == "generateContent" {
			return true
		}
	}
	return false
}

func containsFold(haystack, needle string) bool {
	return strings.Contains(strings.ToLower(haystack), strings.ToLower(needle))
}
