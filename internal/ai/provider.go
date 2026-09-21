// Copyright (c) 2024
// Licensed under the MIT License. See LICENSE file in the root directory.

package ai

import (
	"context"
	"errors"
	"fmt"
	"strings"
)

const (
	KindOpenAI           = "openai"
	KindOpenRouter       = "openrouter"
	KindGemini           = "gemini"
	KindOpenAICompatible = "openai-compatible"
)

const defaultTemperature = 0.7

// Gemini caps an unspecified request at 8192, well below what the model can do.
// OpenAI-compatible endpoints are left alone: max_tokens is deprecated and the
// newer reasoning models reject it.
const geminiMaxOutputTokens = 32768

var ErrNoProvider = errors.New("no AI provider is configured")

func BuiltIn() []string { return []string{KindOpenAI, KindOpenRouter, KindGemini} }

func IsBuiltIn(name string) bool {
	for _, kind := range BuiltIn() {
		if strings.EqualFold(kind, name) {
			return true
		}
	}
	return false
}

func DefaultBaseURL(kind string) string {
	switch kind {
	case KindOpenRouter:
		return openRouterBaseURL
	case KindGemini:
		return geminiBaseURL
	default:
		return openAIBaseURL
	}
}

func DefaultModel(kind string) string {
	switch kind {
	case KindOpenRouter:
		return "openai/gpt-4o-mini"
	case KindGemini:
		return "gemini-2.5-flash"
	default:
		return "gpt-4o-mini"
	}
}

type Settings struct {
	Name string
	Kind string

	APIKey      string
	BaseURL     string
	Model       string
	Temperature float64

	// NoAPIKey suits a model served from this machine.
	NoAPIKey bool
	// LowReasoning asks a reasoning model to think less; the thinking tokens are billed and discarded.
	LowReasoning bool
	Custom       bool
}

func (s Settings) RequiresAPIKey() bool { return !s.NoAPIKey }

func (s Settings) Resolved() Settings {
	if s.Kind == "" {
		if s.Custom {
			s.Kind = KindOpenAICompatible
		} else {
			s.Kind = s.Name
		}
	}
	if !s.Custom {
		if s.BaseURL == "" {
			s.BaseURL = DefaultBaseURL(s.Kind)
		}
		if s.Model == "" {
			s.Model = DefaultModel(s.Kind)
		}
	}
	if s.Temperature == 0 {
		s.Temperature = defaultTemperature
	}
	s.BaseURL = strings.TrimRight(strings.TrimSpace(s.BaseURL), "/")
	return s
}

type Request struct {
	System string
	Prompt string
}

// A non-nil error from emit ends the stream.
type Provider interface {
	Name() string
	Model() string
	Validate() error
	Stream(ctx context.Context, req Request, emit func(chunk string) error) error
}

func New(settings Settings) (Provider, error) {
	settings = settings.Resolved()

	if settings.Custom && settings.BaseURL == "" {
		return nil, fmt.Errorf("a base URL is required for %q", settings.Name)
	}

	switch settings.Kind {
	case KindGemini:
		return &geminiProvider{settings: settings, client: newHTTPClient()}, nil
	case KindOpenAI, KindOpenRouter, KindOpenAICompatible, "":
		return &openAIProvider{settings: settings, client: newHTTPClient()}, nil
	default:
		return nil, fmt.Errorf("unsupported provider type: %s", settings.Kind)
	}
}

func Complete(ctx context.Context, provider Provider, req Request) (string, error) {
	var answer strings.Builder

	err := provider.Stream(ctx, req, func(chunk string) error {
		answer.WriteString(chunk)
		return nil
	})
	if err != nil {
		return "", err
	}

	return answer.String(), nil
}
