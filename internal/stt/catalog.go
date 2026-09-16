// Copyright (c) 2024
// Licensed under the MIT License. See LICENSE file in the root directory.

package stt

import (
	"context"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

// Shipped so the list renders offline and on first run.
//
//go:embed models.json
var embeddedCatalog []byte

// Where a newer list is fetched from, so models can be added between releases.
const CatalogURL = "https://raw.githubusercontent.com/paradoxe35/myscript/main/internal/stt/models.json"

const (
	catalogMaxAge = 24 * time.Hour
	catalogMaxLen = 8 << 20
	modelsDirName = "models"
)

type Catalog struct {
	Version int     `json:"catalog_version"`
	Source  string  `json:"source"`
	Models  []Model `json:"models"`

	origin  string
	fetched time.Time
}

var (
	catalogMu sync.RWMutex
	active    *Catalog
	homeDir   string
)

// Init points the package at the app home: models live in its "models"
// directory and a refreshed catalogue is cached next to them.
func Init(dir string) {
	catalogMu.Lock()
	homeDir = dir
	active = nil
	catalogMu.Unlock()
}

func ModelsDir() string { return filepath.Join(homeDir, modelsDirName) }

func cachePath() string {
	if homeDir == "" {
		return ""
	}
	return filepath.Join(homeDir, "catalog.json")
}

// Models prefers a cached download over the shipped copy.
func Models() *Catalog {
	catalogMu.RLock()
	current := active
	catalogMu.RUnlock()
	if current != nil {
		return current
	}

	catalogMu.Lock()
	defer catalogMu.Unlock()
	if active == nil {
		active = loadBest()
	}
	return active
}

// discoverCustom finds GGUF files not claimed by the catalogue, for fine-tuned
// or community models. Old ggml ".bin" files are a format the engine cannot read.
func discoverCustom() []Model {
	return discoverCustomIn(ModelsDir(), Models().Models)
}

func discoverCustomIn(dir string, published []Model) []Model {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}

	claimed := make(map[string]bool, len(published))
	for _, model := range published {
		claimed[filepath.Base(model.Filename)] = true
	}

	var custom []Model
	for _, entry := range entries {
		if entry.IsDir() || claimed[entry.Name()] || !strings.HasSuffix(entry.Name(), ".gguf") {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			continue
		}
		custom = append(custom, Model{
			ID:        "custom/" + entry.Name(),
			Slug:      entry.Name(),
			Name:      strings.TrimSuffix(entry.Name(), filepath.Ext(entry.Name())),
			Filename:  entry.Name(),
			SizeBytes: info.Size(),
		})
	}
	sort.Slice(custom, func(i, j int) bool { return custom[i].Name < custom[j].Name })
	return custom
}

func loadBest() *Catalog {
	shipped, err := parseCatalog(embeddedCatalog)
	if err != nil {
		panic("stt: embedded catalog is invalid: " + err.Error())
	}
	shipped.origin = "embedded"

	path := cachePath()
	if path == "" {
		return shipped
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return shipped
	}

	cached, err := parseCatalog(data)
	if err != nil {
		slog.Warn("Ignoring an unreadable catalog cache", "error", err)
		return shipped
	}

	// A cache older than what shipped means the app was updated since.
	if cached.Version < shipped.Version {
		return shipped
	}

	cached.origin = "cached"
	if info, err := os.Stat(path); err == nil {
		cached.fetched = info.ModTime()
	}
	return cached
}

func parseCatalog(data []byte) (*Catalog, error) {
	var catalog Catalog
	if err := json.Unmarshal(data, &catalog); err != nil {
		return nil, err
	}
	if len(catalog.Models) == 0 {
		return nil, errors.New("catalog contains no models")
	}
	for _, model := range catalog.Models {
		if !validSHA256(model.SHA256) {
			return nil, fmt.Errorf("model %s has an unusable checksum %q", model.ID, model.SHA256)
		}
	}

	sort.SliceStable(catalog.Models, func(i, j int) bool {
		if catalog.Models[i].Rank != catalog.Models[j].Rank {
			return catalog.Models[i].Rank < catalog.Models[j].Rank
		}
		return catalog.Models[i].AccuracyScore > catalog.Models[j].AccuracyScore
	})
	return &catalog, nil
}

func stale() bool {
	catalog := Models()
	return catalog.origin == "embedded" || time.Since(catalog.fetched) > catalogMaxAge
}

// Refresh replaces the cached list; parsed before writing so a truncated
// download never displaces a working one.
func Refresh(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, CatalogURL, nil)
	if err != nil {
		return err
	}

	resp, err := (&http.Client{Timeout: 20 * time.Second}).Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("catalog fetch returned %s", resp.Status)
	}

	data, err := io.ReadAll(io.LimitReader(resp.Body, catalogMaxLen))
	if err != nil {
		return err
	}

	fetched, err := parseCatalog(data)
	if err != nil {
		return fmt.Errorf("published catalog is unusable: %w", err)
	}

	if path := cachePath(); path != "" {
		if err := os.WriteFile(path, data, 0o644); err != nil {
			slog.Warn("Could not cache the model catalog", "error", err)
		}
	}

	fetched.origin = "cached"
	fetched.fetched = time.Now()

	catalogMu.Lock()
	active = fetched
	catalogMu.Unlock()

	slog.Info("Model catalog refreshed", "models", len(fetched.Models), "version", fetched.Version)
	return nil
}

// RefreshInBackground never blocks startup; failures aren't surfaced since the
// shipped list still works.
func RefreshInBackground() {
	if !stale() {
		return
	}
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		if err := Refresh(ctx); err != nil {
			slog.Info("Keeping the shipped model catalog", "reason", err)
		}
	}()
}

// Catalogue is the published list plus whatever the user dropped into the
// models directory. Copied rather than appended in place: the parsed slice has
// spare capacity that other callers still read.
func Catalogue() []Model {
	published := Models().Models
	custom := discoverCustom()

	all := make([]Model, 0, len(published)+len(custom))
	all = append(all, published...)
	return append(all, custom...)
}

func FindModel(id string) (Model, bool) {
	for _, model := range Catalogue() {
		if model.ID == id {
			return model, true
		}
	}
	return Model{}, false
}
