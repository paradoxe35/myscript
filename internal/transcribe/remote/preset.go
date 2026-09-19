// Copyright (c) 2024
// Licensed under the MIT License. See LICENSE file in the root directory.

// Package remote transcribes through a hosted service. Most speak the OpenAI
// multipart endpoint; Gemini has no equivalent and is asked through generateContent.
package remote

import "strings"

type Protocol string

const (
	ProtocolOpenAI Protocol = "openai"
	ProtocolGemini Protocol = "gemini"
)

const CustomPreset = "custom"

// Preset is a hosted transcription service the settings screen offers.
type Preset struct {
	ID       string
	Name     string
	BaseURL  string
	Models   []string
	KeyHint  string
	Protocol Protocol
}

func (p Preset) Custom() bool { return p.ID == CustomPreset }

var Presets = []Preset{
	{
		ID:       "openai",
		Name:     "OpenAI",
		BaseURL:  "https://api.openai.com/v1",
		Models:   []string{"gpt-4o-mini-transcribe", "gpt-4o-transcribe", "whisper-1"},
		KeyHint:  "platform.openai.com",
		Protocol: ProtocolOpenAI,
	},
	{
		ID:       "groq",
		Name:     "Groq",
		BaseURL:  "https://api.groq.com/openai/v1",
		Models:   []string{"whisper-large-v3-turbo", "whisper-large-v3"},
		KeyHint:  "console.groq.com",
		Protocol: ProtocolOpenAI,
	},
	{
		// Only the chat models: the dedicated speech models answer on a
		// different endpoint this does not speak.
		ID:       "gemini",
		Name:     "Google Gemini",
		BaseURL:  "https://generativelanguage.googleapis.com",
		Models:   []string{"gemini-2.5-flash", "gemini-2.0-flash"},
		KeyHint:  "aistudio.google.com",
		Protocol: ProtocolGemini,
	},
	{
		ID:       CustomPreset,
		Name:     "Custom",
		KeyHint:  "any OpenAI-compatible endpoint",
		Protocol: ProtocolOpenAI,
	},
}

func FindPreset(id string) (Preset, bool) {
	for _, preset := range Presets {
		if strings.EqualFold(preset.ID, id) {
			return preset, true
		}
	}
	return Preset{}, false
}

// DefaultPreset is what a configuration naming nothing falls back to.
func DefaultPreset() Preset { return Presets[0] }

// An unknown id is a configuration naming a preset this build does not have,
// which has been OpenAI-shaped in every case so far.
func protocolFor(id string) Protocol {
	if preset, ok := FindPreset(id); ok && preset.Protocol != "" {
		return preset.Protocol
	}
	return ProtocolOpenAI
}

// Settings is one configured hosted service.
type Settings struct {
	Preset  string
	BaseURL string
	Model   string
	APIKey  string
}

// Resolved fills in whatever the settings screen left blank.
func (s Settings) Resolved() Settings {
	preset, ok := FindPreset(s.Preset)
	if !ok {
		preset = DefaultPreset()
		s.Preset = preset.ID
	}

	if !preset.Custom() {
		// Filled in, not forced: the settings screen locks the field, so a
		// stored endpoint is a deliberate one and is left alone.
		if s.BaseURL == "" {
			s.BaseURL = preset.BaseURL
		}
		if s.Model == "" && len(preset.Models) > 0 {
			s.Model = preset.Models[0]
		}
	}

	s.BaseURL = strings.TrimRight(strings.TrimSpace(s.BaseURL), "/")
	s.Model = strings.TrimSpace(s.Model)
	return s
}

func (s Settings) Ready() bool {
	resolved := s.Resolved()
	return resolved.BaseURL != "" && resolved.Model != ""
}
