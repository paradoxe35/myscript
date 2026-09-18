// Copyright (c) 2024
// Licensed under the MIT License. See LICENSE file in the root directory.

package ai

import (
	"io"
	"net/http"
	"strings"
	"testing"
)

func sseResponse(body string) *http.Response {
	return &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(strings.NewReader(body)),
	}
}

func collectEvents(t *testing.T, body string) []sseEvent {
	t.Helper()

	var events []sseEvent
	err := streamSSE(sseResponse(body), "test", func(event sseEvent) error {
		events = append(events, event)
		return nil
	})
	if err != nil {
		t.Fatalf("streamSSE: %v", err)
	}
	return events
}

func TestStreamSSEReadsNamedEvents(t *testing.T) {
	events := collectEvents(t, "event: first\ndata: one\n\nevent: second\ndata: two\n\n")

	want := []sseEvent{{Name: "first", Data: "one"}, {Name: "second", Data: "two"}}
	if len(events) != len(want) {
		t.Fatalf("got %d events, want %d", len(events), len(want))
	}
	for i, event := range events {
		if event != want[i] {
			t.Errorf("event %d = %+v, want %+v", i, event, want[i])
		}
	}
}

func TestStreamSSEJoinsMultiLineData(t *testing.T) {
	events := collectEvents(t, "data: {\"a\":\ndata: 1}\n\n")

	if len(events) != 1 || events[0].Data != "{\"a\":\n1}" {
		t.Fatalf("got %+v", events)
	}
}

func TestStreamSSEDeliversTrailingEventWithoutBlankLine(t *testing.T) {
	events := collectEvents(t, "data: last\n")

	if len(events) != 1 || events[0].Data != "last" {
		t.Fatalf("got %+v", events)
	}
}

func TestStreamSSESurfacesHTTPErrors(t *testing.T) {
	resp := &http.Response{
		StatusCode: http.StatusUnauthorized,
		Body:       io.NopCloser(strings.NewReader(`{"error":{"message":"bad key"}}`)),
	}

	err := streamSSE(resp, "openai", func(sseEvent) error { return nil })
	if err == nil {
		t.Fatal("expected an error")
	}
	if !strings.Contains(err.Error(), "bad key") {
		t.Errorf("error %q should carry the provider message", err)
	}

	apiErr, ok := err.(*APIError)
	if !ok || apiErr.StatusCode != http.StatusUnauthorized {
		t.Errorf("error should be an APIError carrying the status, got %#v", err)
	}
}

func TestParseAPIErrorFallsBackToStatus(t *testing.T) {
	err := parseAPIError(http.StatusTooManyRequests, nil, "gemini")

	if !strings.Contains(err.Error(), "rate limit") {
		t.Errorf("got %q", err)
	}
}
