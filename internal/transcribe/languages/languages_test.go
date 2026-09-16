package languages

import "testing"

func TestValidateAcceptsOnlyWhisperCodes(t *testing.T) {
	if err := Validate("fr"); err != nil {
		t.Errorf("fr should be valid: %v", err)
	}
	if err := Validate("he"); err != nil {
		t.Errorf("Hebrew is 'he' for Whisper: %v", err)
	}
	if err := Validate("iw"); err == nil {
		t.Error("the legacy 'iw' code is not what Whisper takes")
	}
	if err := Validate("xx"); err != ErrInvalid {
		t.Errorf("unknown code returned %v, want ErrInvalid", err)
	}
}

func TestWhisperListIsNamedAndUnique(t *testing.T) {
	if len(Whisper) != 60 || Whisper[0].Code != "en" {
		t.Fatalf("Whisper list has %d entries starting with %s", len(Whisper), Whisper[0].Code)
	}
	seen := map[string]bool{}
	for _, language := range Whisper {
		if seen[language.Code] {
			t.Errorf("duplicate code %s", language.Code)
		}
		seen[language.Code] = true
		if language.Name == language.Code {
			t.Errorf("%s has no English name", language.Code)
		}
	}
}

func TestNameFallsBackToTheCode(t *testing.T) {
	if got := Name("FR"); got != "French" {
		t.Errorf("Name(FR) = %q", got)
	}
	if got := Name("zzz"); got != "zzz" {
		t.Errorf("unknown code should display as-is, got %q", got)
	}
}

func TestNamedPutsEnglishFirstThenAlphabetical(t *testing.T) {
	named := Named([]string{"zh", "fr", "en", "de"})
	want := []string{"en", "zh", "fr", "de"}
	for i, language := range named {
		if language.Code != want[i] {
			t.Fatalf("order = %v, want %v", named, want)
		}
	}
	if named[1].Name != "Chinese" {
		t.Errorf("zh should be named Chinese, got %q", named[1].Name)
	}
}
