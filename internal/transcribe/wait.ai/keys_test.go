// Copyright (c) 2024
// Licensed under the MIT License. See LICENSE file in the root directory.

package witai

import "testing"

func withKeys(t *testing.T, raw string) {
	t.Helper()

	rawKeys = raw
	parseKeys = newParseKeys()

	t.Cleanup(func() {
		rawKeys = ""
		parseKeys = newParseKeys()
	})
}

func TestKeysAreReadPerLanguage(t *testing.T) {
	withKeys(t, `[{"key":"ABC","lang":"en"},{"key":"DEF","lang":"fr"}]`)

	if !Available() {
		t.Fatal("embedded keys should be available")
	}
	if token, ok := Token("fr"); !ok || token != "DEF" {
		t.Errorf("got %q, %v", token, ok)
	}
	if _, ok := Token("de"); ok {
		t.Error("a language without a key should report none")
	}
}

func TestInvalidKeysBehaveAsNone(t *testing.T) {
	withKeys(t, "not json")

	if Available() {
		t.Fatal("unparsable keys must behave as no keys")
	}
}

func TestNoKeysEmbedded(t *testing.T) {
	withKeys(t, "")

	if Available() {
		t.Fatal("expected no keys")
	}
	if len(GetSupportedLanguages()) != 0 {
		t.Error("no keys means no supported languages")
	}
}

func TestIncompleteKeysAreDropped(t *testing.T) {
	withKeys(t, `[{"key":"ABC","lang":"en"},{"key":"","lang":"fr"},{"key":"GHI","lang":""}]`)

	languages := GetSupportedLanguages()
	if len(languages) != 1 || languages[0].Code != "en" {
		t.Fatalf("got %+v", languages)
	}
}
