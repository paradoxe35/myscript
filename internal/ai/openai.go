// Copyright (c) 2024
// Licensed under the MIT License. See LICENSE file in the root directory.

package ai

import (
	"context"
	"encoding/json"
	"fmt"
)

const openAIBaseURL = "https://api.openai.com/v1"

type openAIProvider struct {
	settings Settings
	client   httpDoer
}

type openAIMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type openRouterReasoning struct {
	Effort  string `json:"effort"`
	Exclude bool   `json:"exclude"`
}

type openAIRequest struct {
	Model           string               `json:"model"`
	Messages        []openAIMessage      `json:"messages"`
	Temperature     float64              `json:"temperature"`
	Stream          bool                 `json:"stream"`
	ReasoningEffort string               `json:"reasoning_effort,omitempty"`
	Reasoning       *openRouterReasoning `json:"reasoning,omitempty"`
}

func (p *openAIProvider) Name() string  { return p.settings.Name }
func (p *openAIProvider) Model() string { return p.settings.Model }

func (p *openAIProvider) Validate() error {
	if p.settings.BaseURL == "" {
		return fmt.Errorf("%s needs a base URL", p.settings.Name)
	}
	if p.settings.Model == "" {
		return fmt.Errorf("%s needs a model", p.settings.Name)
	}
	if p.settings.RequiresAPIKey() && p.settings.APIKey == "" {
		return fmt.Errorf("%s needs an API key", p.settings.Name)
	}
	return nil
}

func (p *openAIProvider) Stream(ctx context.Context, req Request, emit func(string) error) error {
	if err := p.Validate(); err != nil {
		return err
	}

	messages := []openAIMessage{
		{Role: "system", Content: req.System},
		{Role: "user", Content: req.Prompt},
	}

	return withReasoningFallback(p.settings, func(lowReasoning bool) error {
		body := openAIRequest{
			Model:       p.settings.Model,
			Messages:    messages,
			Temperature: p.settings.Temperature,
			Stream:      true,
		}

		if lowReasoning {
			// "low" rather than "none": some models refuse to switch reasoning off entirely.
			switch reasoningStyleOf(p.settings.BaseURL) {
			case reasoningOpenRouter:
				body.Reasoning = &openRouterReasoning{Effort: "low", Exclude: true}
			default:
				body.ReasoningEffort = "low"
			}
		}

		return p.send(ctx, body, emit)
	})
}

func (p *openAIProvider) send(ctx context.Context, body openAIRequest, emit func(string) error) error {
	headers := map[string]string{"Accept": "text/event-stream"}
	// A local runtime takes no credentials, and some reject an empty bearer.
	if p.settings.APIKey != "" {
		headers["Authorization"] = "Bearer " + p.settings.APIKey
	}

	resp, err := postJSON(ctx, p.client, p.settings.BaseURL+"/chat/completions", headers, body)
	if err != nil {
		return err
	}

	return streamSSE(resp, p.settings.Name, func(event sseEvent) error {
		if event.Data == "[DONE]" {
			return nil
		}

		var chunk struct {
			Choices []struct {
				Delta struct {
					Content string `json:"content"`
				} `json:"delta"`
			} `json:"choices"`
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

		for _, choice := range chunk.Choices {
			if choice.Delta.Content == "" {
				continue
			}
			if err := emit(choice.Delta.Content); err != nil {
				return err
			}
		}

		return nil
	})
}
