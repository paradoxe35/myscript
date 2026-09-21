// Copyright (c) 2024
// Licensed under the MIT License. See LICENSE file in the root directory.

// Package stt owns local speech-to-text: the model catalogue, the model store,
// how models fit this machine, and the service that turns a take into text.
package stt

import "fmt"

// Field order is the file's key order so the generator writes the same shape
// it reads.
type Model struct {
	ID          string `json:"id"`
	Slug        string `json:"slug"`
	Name        string `json:"name"`
	Description string `json:"description"`

	Repo     string `json:"repo"`
	Revision string `json:"revision"`
	Filename string `json:"filename"`
	Quant    string `json:"quant"`

	SizeBytes int64  `json:"size_bytes"`
	SHA256    string `json:"sha256"`

	Languages      []string `json:"languages"`
	License        string   `json:"license"`
	Translate      bool     `json:"translate"`
	Streaming      bool     `json:"streaming"`
	LanguageDetect bool     `json:"language_detect"`

	// WordErrorRate is a percentage; nil when unpublished, which is not a perfect zero.
	WordErrorRate  *float64 `json:"word_error_rate"`
	RealtimeFactor float64  `json:"realtime_factor"`
	SpeedScore     float64  `json:"speed_score"`
	AccuracyScore  float64  `json:"accuracy_score"`

	Recommended bool `json:"recommended"`
	Rank        int  `json:"rank"`
}

// Pins the revision so the repo cannot move under the checksum.
func (m Model) DownloadURL() string {
	return fmt.Sprintf("https://huggingface.co/%s/resolve/%s/%s", m.Repo, m.Revision, m.Filename)
}

func (m Model) SizeMB() float64 { return float64(m.SizeBytes) / (1 << 20) }

func (m Model) Speaks(code string) bool {
	for _, language := range m.Languages {
		if language == code {
			return true
		}
	}
	return false
}

// A model that cannot detect is never left blank: the library would assume English.
func (m Model) TranscribeLanguage(preferred string) string {
	if preferred != "" && m.Speaks(preferred) {
		return preferred
	}
	if m.LanguageDetect {
		return ""
	}
	if len(m.Languages) > 0 {
		return m.Languages[0]
	}
	return ""
}
