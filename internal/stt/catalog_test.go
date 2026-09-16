package stt

import (
	"os"
	"path/filepath"
	"testing"
)

func TestEmbeddedCatalogIsComplete(t *testing.T) {
	models := Models().Models
	if len(models) < 20 {
		t.Fatalf("catalog holds %d models, expected the generated list", len(models))
	}

	seen := make(map[string]bool, len(models))
	for _, m := range models {
		switch {
		case m.ID == "":
			t.Errorf("%s has no id", m.Slug)
		case m.SHA256 == "":
			t.Errorf("%s has no checksum; it could not be verified after download", m.Slug)
		case m.SizeBytes <= 0:
			t.Errorf("%s has no size; resume and progress both need it", m.Slug)
		case m.Revision == "":
			t.Errorf("%s has no pinned revision", m.Slug)
		case len(m.Languages) == 0:
			t.Errorf("%s declares no languages; the picker could not offer any", m.Slug)
		}
		if seen[m.ID] {
			t.Errorf("duplicate model id %s", m.ID)
		}
		seen[m.ID] = true
	}
}

func TestCatalogRanksRecommendedFirst(t *testing.T) {
	first := Models().Models[0]
	if !first.Recommended || first.Rank != 1 {
		t.Errorf("first model %s is not the top-ranked recommendation", first.Slug)
	}
}

func TestDiscoverCustomIgnoresOldGgmlFiles(t *testing.T) {
	dir := t.TempDir()
	write := func(name string, size int64) {
		if err := os.WriteFile(filepath.Join(dir, name), make([]byte, size), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write("my-finetune.gguf", 10)
	write("ggml-base.bin", 20)
	write("notes.txt", 5)
	if err := os.Mkdir(filepath.Join(dir, "nested.gguf"), 0o755); err != nil {
		t.Fatal(err)
	}

	custom := discoverCustomIn(dir, Models().Models)
	if len(custom) != 1 {
		t.Fatalf("discovered %d models, want only the gguf: %+v", len(custom), custom)
	}
	if custom[0].ID != "custom/my-finetune.gguf" || custom[0].Name != "my-finetune" || custom[0].SizeBytes != 10 {
		t.Errorf("custom model fields wrong: %+v", custom[0])
	}
}

func TestDiscoverCustomSkipsCatalogEntries(t *testing.T) {
	dir := t.TempDir()
	shipped := Models().Models[0]
	if err := os.WriteFile(filepath.Join(dir, shipped.Filename), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if got := discoverCustomIn(dir, Models().Models); len(got) != 0 {
		t.Errorf("a file claimed by the catalog was rediscovered: %+v", got)
	}
}

// The store keeps only the base name, so a catalogue filename with a directory
// component must be claimed by its base name too or it is listed twice.
func TestDiscoverCustomClaimsSlashedFilenamesByBase(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "bundled.gguf"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	published := []Model{{ID: "org/bundled", Filename: "bundle/bundled.gguf"}}
	if got := discoverCustomIn(dir, published); len(got) != 0 {
		t.Errorf("a bundled catalogue file was rediscovered as custom: %+v", got)
	}
}

func TestParseCatalogRejectsUnusableChecksums(t *testing.T) {
	for _, sum := range []string{"", "abc", "zz" + digest(nil)[2:]} {
		data := []byte(`{"catalog_version":1,"models":[{"id":"a","sha256":"` + sum + `"}]}`)
		if _, err := parseCatalog(data); err == nil {
			t.Errorf("checksum %q was accepted", sum)
		}
	}
}

func TestDownloadURLPinsRevision(t *testing.T) {
	model := Model{Repo: "org/repo", Revision: "abc123", Filename: "m.gguf"}
	want := "https://huggingface.co/org/repo/resolve/abc123/m.gguf"
	if got := model.DownloadURL(); got != want {
		t.Errorf("DownloadURL() = %q, want %q", got, want)
	}
}

func TestParseCatalogRejectsEmptyOrBroken(t *testing.T) {
	if _, err := parseCatalog([]byte(`{"catalog_version":1,"models":[]}`)); err == nil {
		t.Error("an empty catalog must be rejected, not adopted")
	}
	if _, err := parseCatalog([]byte(`not json`)); err == nil {
		t.Error("unparseable data must be rejected")
	}
}

func TestParseCatalogOrdersByRankThenAccuracy(t *testing.T) {
	sum := digest(nil)
	catalog, err := parseCatalog([]byte(`{"catalog_version":1,"models":[
		{"id":"c","rank":999,"accuracy_score":0.5,"sha256":"` + sum + `"},
		{"id":"a","rank":1,"accuracy_score":0.1,"sha256":"` + sum + `"},
		{"id":"b","rank":999,"accuracy_score":0.9,"sha256":"` + sum + `"}]}`))
	if err != nil {
		t.Fatal(err)
	}
	got := []string{catalog.Models[0].ID, catalog.Models[1].ID, catalog.Models[2].ID}
	if got[0] != "a" || got[1] != "b" || got[2] != "c" {
		t.Errorf("order = %v, want a, b, c", got)
	}
}

func TestFindModelMatchesByID(t *testing.T) {
	first := Catalogue()[0]
	found, ok := FindModel(first.ID)
	if !ok || found.ID != first.ID {
		t.Fatalf("FindModel(%q) did not round-trip", first.ID)
	}
	if _, ok := FindModel("nothing/here"); ok {
		t.Error("an unknown id should not resolve")
	}
}

func TestCatalogueDoesNotAliasTheCatalog(t *testing.T) {
	published := Models().Models
	combined := Catalogue()
	if len(combined) < len(published) {
		t.Fatalf("Catalogue() dropped entries: %d < %d", len(combined), len(published))
	}
	if &combined[0] == &published[0] {
		t.Error("Catalogue() shares its backing array with the published catalog")
	}
}

func TestTranscribeLanguage(t *testing.T) {
	detecting := Model{Languages: []string{"en", "fr"}, LanguageDetect: true}
	fixed := Model{Languages: []string{"en", "fr"}}
	single := Model{Languages: []string{"en"}}

	cases := []struct {
		name      string
		model     Model
		preferred string
		want      string
	}{
		{"spoken choice wins", detecting, "fr", "fr"},
		{"detecting model is left to detect", detecting, "", ""},
		{"unspoken language falls back to detection", detecting, "de", ""},
		{"fixed model never blank", fixed, "", "en"},
		{"fixed model refuses what it cannot speak", fixed, "de", "en"},
		{"single language model", single, "", "en"},
	}
	for _, c := range cases {
		if got := c.model.TranscribeLanguage(c.preferred); got != c.want {
			t.Errorf("%s: got %q, want %q", c.name, got, c.want)
		}
	}
}
