// Copyright (c) 2024
// Licensed under the MIT License. See LICENSE file in the root directory.

package ai

import (
	"context"
	"encoding/json"
	"fmt"
)

const geminiBaseURL = "https://generativelanguage.googleapis.com"

type geminiProvider struct {
	settings Settings
	client   httpDoer
}

type geminiPart struct {
	Text string `json:"text"`
}

type geminiContent struct {
	Parts []geminiPart `json:"parts"`
	Role  string       `json:"role,omitempty"`
}

type geminiThinking struct {
	ThinkingBudget int `json:"thinkingBudget"`
}

type geminiGeneration struct {
	Temperature    float64         `json:"temperature"`
	ThinkingConfig *geminiThinking `json:"thinkingConfig,omitempty"`
}

type geminiRequest struct {
	Contents          []geminiContent  `json:"contents"`
	SystemInstruction *geminiContent   `json:"systemInstruction,omitempty"`
	GenerationConfig  geminiGeneration `json:"generationConfig"`
}

func (p *geminiProvider) Name() string  { return p.settings.Name }
func (p *geminiProvider) Model() string { return p.settings.Model }

func (p *geminiProvider) Validate() error {
	if p.settings.APIKey == "" {
		return fmt.Errorf("%s needs an API key", p.settings.Name)
	}
	if p.settings.BaseURL == "" {
		return fmt.Errorf("%s needs a base URL", p.settings.Name)
	}
	if p.settings.Model == "" {
		return fmt.Errorf("%s needs a model", p.settings.Name)
	}
	return nil
}

func (p *geminiProvider) Stream(ctx context.Context, req Request, emit func(string) error) error {
	if err := p.Validate(); err != nil {
		return err
	}

	body := geminiRequest{
		Contents:         []geminiContent{{Role: "user", Parts: []geminiPart{{Text: req.Prompt}}}},
		GenerationConfig: geminiGeneration{Temperature: p.settings.Temperature},
	}
	if req.System != "" {
		body.SystemInstruction = &geminiContent{Parts: []geminiPart{{Text: req.System}}}
	}

	return withReasoningFallback(p.settings, func(lowReasoning bool) error {
		if lowReasoning {
			body.GenerationConfig.ThinkingConfig = &geminiThinking{ThinkingBudget: 0}
		} else {
			body.GenerationConfig.ThinkingConfig = nil
		}
		return p.send(ctx, body, emit)
	})
}

func (p *geminiProvider) send(ctx context.Context, body geminiRequest, emit func(string) error) error {
	url := fmt.Sprintf("%s/v1beta/models/%s:streamGenerateContent?alt=sse", p.settings.BaseURL, p.settings.Model)
	headers := map[string]string{
		"x-goog-api-key": p.settings.APIKey,
		"Accept":         "text/event-stream",
	}

	resp, err := postJSON(ctx, p.client, url, headers, body)
	if err != nil {
		return err
	}

	return streamSSE(resp, p.settings.Name, func(event sseEvent) error {
		var chunk struct {
			Candidates []struct {
				Content geminiContent `json:"content"`
			} `json:"candidates"`
			Error *struct {
				Message string `json:"message"`
			} `json:"error"`
		}

		if err := json.Unmarshal([]byte(event.Data), &chunk); err != nil {
			return decodeError(err, []byte(event.Data), p.settings.Name)
		}
		if chunk.Error != nil {
			return fmt.Errorf("%s API error: %s", p.settings.Name, chunk.Error.Message)
		}

		for _, candidate := range chunk.Candidates {
			for _, part := range candidate.Content.Parts {
				if part.Text == "" {
					continue
				}
				if err := emit(part.Text); err != nil {
					return err
				}
			}
		}

		return nil
	})
}
