// Copyright (c) 2024
// Licensed under the MIT License. See LICENSE file in the root directory.

package remote

import "testing"

func TestResolvedFillsInAKnownService(t *testing.T) {
	settings := Settings{Preset: "groq"}.Resolved()

	if settings.BaseURL != "https://api.groq.com/openai/v1" {
		t.Errorf("BaseURL = %q", settings.BaseURL)
	}
	if settings.Model != "whisper-large-v3-turbo" {
		t.Errorf("Model = %q", settings.Model)
	}
}

func TestResolvedKeepsAStoredEndpoint(t *testing.T) {
	settings := Settings{Preset: "openai", BaseURL: "https://proxy.internal/v1/"}.Resolved()

	if settings.BaseURL != "https://proxy.internal/v1" {
		t.Errorf("BaseURL = %q, a stored endpoint should survive with its slash trimmed", settings.BaseURL)
	}
}

func TestResolvedFallsBackForAnUnknownService(t *testing.T) {
	settings := Settings{Preset: "retired"}.Resolved()

	if settings.Preset != DefaultPreset().ID {
		t.Errorf("Preset = %q", settings.Preset)
	}
}

func TestCustomKeepsItsOwnEndpointAndModel(t *testing.T) {
	settings := Settings{Preset: CustomPreset, BaseURL: "http://localhost:8080/v1"}.Resolved()

	if settings.BaseURL != "http://localhost:8080/v1" {
		t.Errorf("BaseURL = %q", settings.BaseURL)
	}
	if settings.Model != "" {
		t.Errorf("Model = %q, a custom service has no default", settings.Model)
	}
}

func TestReadyNeedsAnEndpointAndModel(t *testing.T) {
	if (Settings{Preset: CustomPreset, BaseURL: "http://x"}).Ready() {
		t.Error("a custom service without a model is not ready")
	}
	if !(Settings{Preset: "openai"}).Ready() {
		t.Error("a known service is ready on its defaults")
	}
}

func TestLanguagesFor(t *testing.T) {
	if got := LanguagesFor("gemini", "gemini-2.5-flash"); len(got) < 100 {
		t.Errorf("Gemini is asked in prose and accepts any language, got %d", len(got))
	}
	if got := LanguagesFor("openai", "whisper-1"); len(got) == 0 {
		t.Error("a known service should offer its list")
	}
	if got := LanguagesFor(CustomPreset, "some-local-model"); got != nil {
		t.Errorf("an unknown model should let a code be typed, got %d entries", len(got))
	}
	if got := LanguagesFor(CustomPreset, "faster-whisper-large"); len(got) == 0 {
		t.Error("a whisper-like custom model should suggest Whisper's codes")
	}
}
