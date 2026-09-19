// Copyright (c) 2024
// Licensed under the MIT License. See LICENSE file in the root directory.

package remote

import (
	"myscript/internal/transcribe/languages"
	"strings"
)

// LanguagesFor is what a service and model accept. Nil means the list is not
// known, so the caller should let a code be typed rather than show a wrong list.
func LanguagesFor(preset, model string) []languages.Language {
	if protocolFor(preset) == ProtocolGemini {
		// Asked for in prose, so anything the model reads is fair game.
		return languages.All()
	}

	if known, ok := FindPreset(preset); ok && !known.Custom() {
		return languages.Whisper
	}

	// A custom endpoint is a Whisper server often enough to be worth offering.
	if whisperLike(model) {
		return languages.Whisper
	}
	return nil
}

func whisperLike(model string) bool {
	model = strings.ToLower(model)
	return strings.Contains(model, "whisper") || strings.HasPrefix(model, "gpt-")
}
