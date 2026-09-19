// Copyright (c) 2024
// Licensed under the MIT License. See LICENSE file in the root directory.

package ai

import "testing"

func TestResolvedFillsBuiltInDefaults(t *testing.T) {
	settings := Settings{Name: KindOpenRouter}.Resolved()

	if settings.Kind != KindOpenRouter {
		t.Errorf("Kind = %q", settings.Kind)
	}
	if settings.BaseURL != openRouterBaseURL {
		t.Errorf("BaseURL = %q", settings.BaseURL)
	}
	if settings.Model != DefaultModel(KindOpenRouter) {
		t.Errorf("Model = %q", settings.Model)
	}
	if settings.Temperature != defaultTemperature {
		t.Errorf("Temperature = %v", settings.Temperature)
	}
}

func TestResolvedLeavesCustomProvidersAlone(t *testing.T) {
	settings := Settings{Name: "local", Custom: true, BaseURL: "http://localhost:1234/v1/"}.Resolved()

	if settings.Kind != KindOpenAICompatible {
		t.Errorf("Kind = %q", settings.Kind)
	}
	if settings.BaseURL != "http://localhost:1234/v1" {
		t.Errorf("BaseURL = %q, trailing slash should be trimmed", settings.BaseURL)
	}
	if settings.Model != "" {
		t.Errorf("Model = %q, a custom provider has no default", settings.Model)
	}
}

func TestNewSelectsTheProviderForTheKind(t *testing.T) {
	cases := map[string]any{
		KindOpenAI:           &openAIProvider{},
		KindOpenRouter:       &openAIProvider{},
		KindGemini:           &geminiProvider{},
		KindOpenAICompatible: &openAIProvider{},
	}

	for kind, want := range cases {
		provider, err := New(Settings{Name: kind, Kind: kind, BaseURL: "https://example.test"})
		if err != nil {
			t.Fatalf("%s: %v", kind, err)
		}
		if got, wantType := typeName(provider), typeName(want); got != wantType {
			t.Errorf("%s built %s, want %s", kind, got, wantType)
		}
	}
}

func TestNewRejectsUnknownKind(t *testing.T) {
	if _, err := New(Settings{Name: "x", Kind: "telepathy"}); err == nil {
		t.Fatal("expected an error")
	}
}

func TestNewRequiresBaseURLForCustomProviders(t *testing.T) {
	if _, err := New(Settings{Name: "local", Custom: true}); err == nil {
		t.Fatal("expected an error")
	}
}

func TestRequiresAPIKey(t *testing.T) {
	if !(Settings{}).RequiresAPIKey() {
		t.Error("a provider requires a key unless it opts out")
	}
	if (Settings{NoAPIKey: true}).RequiresAPIKey() {
		t.Error("NoAPIKey should waive the requirement")
	}
}

func typeName(value any) string {
	switch value.(type) {
	case *openAIProvider:
		return "openai"
	case *geminiProvider:
		return "gemini"
	default:
		return "unknown"
	}
}
