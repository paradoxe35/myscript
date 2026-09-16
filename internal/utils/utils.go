// Copyright (c) 2024
// Licensed under the MIT License. See LICENSE file in the root directory.

package utils

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"
)

func MeasureExec(name string) func() {
	start := time.Now()

	return func() {
		slog.Debug(fmt.Sprintf("%s took %v", name, time.Since(start)))
	}
}

func IsDevMode() bool {
	return strings.Contains(os.Args[0], "-dev")
}

func HasInternet() bool {
	client := http.Client{Timeout: time.Duration(5000 * time.Millisecond)}
	if _, err := client.Get("http://clients3.google.com/generate_204"); err != nil {
		return false
	}
	return true
}

func IsAOlderThanBByOneWeek(a, b time.Time) bool {
	duration := b.Sub(a)
	// Check if duration is >= 7 days
	return duration >= 7*24*time.Hour
}

func UniqueStrings(input []string) []string {
	seen := make(map[string]bool)
	result := []string{}

	for _, value := range input {
		if !seen[value] {
			seen[value] = true
			result = append(result, value)
		}
	}
	return result
}
