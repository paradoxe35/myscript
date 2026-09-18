// Copyright (c) 2024
// Licensed under the MIT License. See LICENSE file in the root directory.

package ai

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const (
	requestTimeout = 2 * time.Minute
	maxErrorBody   = 4 << 10
	sseBufferSize  = 1 << 20
)

type httpDoer interface {
	Do(req *http.Request) (*http.Response, error)
}

func newHTTPClient() *http.Client {
	return &http.Client{Timeout: requestTimeout}
}

type sseEvent struct {
	Name string
	Data string
}

func postJSON(ctx context.Context, client httpDoer, url string, headers map[string]string, payload any) (*http.Response, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to encode request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	for name, value := range headers {
		req.Header.Set(name, value)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to reach the provider: %w", err)
	}

	return resp, nil
}

func streamSSE(resp *http.Response, provider string, handle func(sseEvent) error) error {
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, maxErrorBody))
		return parseAPIError(resp.StatusCode, body, provider)
	}

	scanner := bufio.NewScanner(resp.Body)
	scanner.Buffer(make([]byte, 0, 64<<10), sseBufferSize)

	var event sseEvent
	for scanner.Scan() {
		line := strings.TrimRight(scanner.Text(), "\r")

		switch {
		case line == "":
			if event.Data == "" {
				continue
			}
			if err := handle(event); err != nil {
				return err
			}
			event = sseEvent{}

		case strings.HasPrefix(line, "event:"):
			event.Name = strings.TrimSpace(strings.TrimPrefix(line, "event:"))

		case strings.HasPrefix(line, "data:"):
			chunk := strings.TrimSpace(strings.TrimPrefix(line, "data:"))
			if event.Data == "" {
				event.Data = chunk
			} else {
				event.Data += "\n" + chunk
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("the stream ended unexpectedly: %w", err)
	}

	if event.Data != "" {
		return handle(event)
	}

	return nil
}
