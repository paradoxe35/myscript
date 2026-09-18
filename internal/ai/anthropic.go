// Copyright (c) 2024
// Licensed under the MIT License. See LICENSE file in the root directory.

package ai

import (
	"context"
	"encoding/json"
	"fmt"
)

const (
	anthropicBaseURL  = "https://api.anthropic.com"
	anthropicVersion  = "2023-06-01"
	anthropicMaxToken = 4096
)

type anthropicProvider struct {
	settings Settings
	client   httpDoer
}

type anthropicMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type anthropicRequest struct {
	Model       string             `json:"model"`
	Messages    []anthropicMessage `json:"messages"`
	MaxTokens   int                `json:"max_tokens"`
	System      string             `json:"system,omitempty"`
	Temperature float64            `json:"temperature"`
	Stream      bool               `json:"stream"`
}

func (p *anthropicProvider) Name() string  { return p.settings.Name }
func (p *anthropicProvider) Model() string { return p.settings.Model }

func (p *anthropicProvider) Validate() error {
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

func (p *anthropicProvider) Stream(ctx context.Context, req Request, emit func(string) error) error {
	if err := p.Validate(); err != nil {
		return err
	}

	body := anthropicRequest{
		Model:       p.settings.Model,
		Messages:    []anthropicMessage{{Role: "user", Content: req.Prompt}},
		MaxTokens:   anthropicMaxToken,
		System:      req.System,
		Temperature: p.settings.Temperature,
		Stream:      true,
	}

	headers := map[string]string{
		"x-api-key":         p.settings.APIKey,
		"anthropic-version": anthropicVersion,
		"Accept":            "text/event-stream",
	}

	resp, err := postJSON(ctx, p.client, p.settings.BaseURL+"/v1/messages", headers, body)
	if err != nil {
		return err
	}

	return streamSSE(resp, p.settings.Name, func(event sseEvent) error {
		switch event.Name {
		case "content_block_delta":
			var delta struct {
				Delta struct {
					Type string `json:"type"`
					Text string `json:"text"`
				} `json:"delta"`
			}
			if err := json.Unmarshal([]byte(event.Data), &delta); err != nil {
				return decodeError(err, []byte(event.Data), p.settings.Name)
			}
			if delta.Delta.Type != "text_delta" || delta.Delta.Text == "" {
				return nil
			}
			return emit(delta.Delta.Text)

		case "error":
			if message := messageFromBody([]byte(event.Data)); message != "" {
				return fmt.Errorf("%s API error: %s", p.settings.Name, message)
			}
			return fmt.Errorf("%s reported an error mid-stream", p.settings.Name)
		}

		return nil
	})
}
