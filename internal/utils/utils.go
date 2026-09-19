// Copyright (c) 2024
// Licensed under the MIT License. See LICENSE file in the root directory.

package utils

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

func MeasureExec(name string) func() {
	start := time.Now()

	return func() {
		slog.Debug(fmt.Sprintf("%s took %v", name, time.Since(start)))
	}
}

// IsDevMode reports a `wails dev` session. The build tag is authoritative; the
// binary name is kept as a fallback for anything that builds without it.
func IsDevMode() bool {
	return devBuild || strings.Contains(os.Args[0], "-dev")
}

const (
	connectivityProbe   = "http://clients3.google.com/generate_204"
	connectivityTimeout = 5 * time.Second
	// Being online rarely changes from one second to the next, so the answer is
	// reused; going offline is rechecked sooner so the app recovers quickly.
	onlineTTL  = 30 * time.Second
	offlineTTL = 5 * time.Second
)

var connectivity struct {
	mu      sync.Mutex
	online  bool
	checked time.Time
}

// HasInternet answers from a short-lived cache, so callers on a timer (the sync
// scheduler ticks every 10s) do not turn a liveness check into a stream of
// requests.
func HasInternet() bool {
	connectivity.mu.Lock()
	defer connectivity.mu.Unlock()

	ttl := offlineTTL
	if connectivity.online {
		ttl = onlineTTL
	}
	if !connectivity.checked.IsZero() && time.Since(connectivity.checked) < ttl {
		return connectivity.online
	}

	connectivity.online = probeInternet()
	connectivity.checked = time.Now()
	return connectivity.online
}

func probeInternet() bool {
	ctx, cancel := context.WithTimeout(context.Background(), connectivityTimeout)
	defer cancel()

	// GET, not HEAD: the endpoint answers 400 to HEAD, which leaves the pooled
	// connection in a state the next request complains about.
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, connectivityProbe, nil)
	if err != nil {
		return false
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	// Drain so the connection can be reused instead of being torn down.
	io.Copy(io.Discard, io.LimitReader(resp.Body, 1<<10))
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
