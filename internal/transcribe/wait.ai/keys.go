// Copyright (c) 2024
// Licensed under the MIT License. See LICENSE file in the root directory.

package witai

import (
	"encoding/json"
	"log/slog"
	"myscript/internal/transcribe/languages"
	"sync"
)

// Set by the generated keys_embedded.go; empty means no keys were embedded.
var rawKeys = ""

// A Wit.ai app serves one language, so there is a key per supported language.
type ApiKey struct {
	Key      string `json:"key"`
	Language string `json:"lang"`
}

// A function so tests can rebind rawKeys and get a fresh cache.
func newParseKeys() func() []ApiKey {
	return sync.OnceValue(func() []ApiKey {
		if rawKeys == "" {
			return nil
		}

		var keys []ApiKey
		if err := json.Unmarshal([]byte(rawKeys), &keys); err != nil {
			slog.Warn("Could not parse the embedded Wit.ai keys", "error", err)
			return nil
		}

		valid := make([]ApiKey, 0, len(keys))
		for _, key := range keys {
			if key.Key != "" && key.Language != "" {
				valid = append(valid, key)
			}
		}
		return valid
	})
}

var parseKeys = newParseKeys()

func Available() bool {
	return len(parseKeys()) > 0
}

func Token(language string) (string, bool) {
	for _, key := range parseKeys() {
		if key.Language == language {
			return key.Key, true
		}
	}
	return "", false
}

// Derived from the embedded keys, so adding one is all a new language needs.
func GetSupportedLanguages() []languages.Language {
	keys := parseKeys()

	codes := make([]string, 0, len(keys))
	for _, key := range keys {
		codes = append(codes, key.Language)
	}

	return languages.Named(codes)
}
