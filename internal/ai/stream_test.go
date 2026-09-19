// Copyright (c) 2024
// Licensed under the MIT License. See LICENSE file in the root directory.

package ai

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type capturedRequest struct {
	path    string
	query   string
	headers http.Header
	body    map[string]any
}

func fakeProvider(t *testing.T, stream string, captured *[]capturedRequest, status ...int) *httptest.Server {
	t.Helper()

	call := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw, _ := io.ReadAll(r.Body)

		var body map[string]any
		json.Unmarshal(raw, &body)
		*captured = append(*captured, capturedRequest{
			path:    r.URL.Path,
			query:   r.URL.RawQuery,
			headers: r.Header.Clone(),
			body:    body,
		})

		if call < len(status) && status[call] != http.StatusOK {
			code := status[call]
			call++
			w.WriteHeader(code)
			io.WriteString(w, `{"error":{"message":"unsupported parameter"}}`)
			return
		}
		call++

		w.Header().Set("Content-Type", "text/event-stream")
		io.WriteString(w, stream)
	}))

	t.Cleanup(server.Close)
	return server
}

func streamText(t *testing.T, provider Provider) string {
	t.Helper()

	text, err := Complete(context.Background(), provider, Request{System: "be brief", Prompt: "hello"})
	if err != nil {
		t.Fatalf("stream: %v", err)
	}
	return text
}

func TestOpenAIStreamsDeltas(t *testing.T) {
	var requests []capturedRequest
	server := fakeProvider(t, `data: {"choices":[{"delta":{"content":"Hello"}}]}

data: {"choices":[{"delta":{"content":", world"}}]}

data: [DONE]

`, &requests)

	provider, err := New(Settings{Name: "openai", Kind: KindOpenAI, APIKey: "sk-test", BaseURL: server.URL, Model: "gpt-4o-mini"})
	if err != nil {
		t.Fatal(err)
	}

	if got := streamText(t, provider); got != "Hello, world" {
		t.Errorf("got %q", got)
	}
	if requests[0].path != "/chat/completions" {
		t.Errorf("path = %q", requests[0].path)
	}
	if got := requests[0].headers.Get("Authorization"); got != "Bearer sk-test" {
		t.Errorf("Authorization = %q", got)
	}
	if requests[0].body["stream"] != true {
		t.Error("the request should ask for a stream")
	}
}

func TestOpenAIOmitsAuthorizationWithoutAKey(t *testing.T) {
	var requests []capturedRequest
	server := fakeProvider(t, "data: [DONE]\n\n", &requests)

	provider, _ := New(Settings{Name: "local", Kind: KindOpenAICompatible, Custom: true, NoAPIKey: true, BaseURL: server.URL, Model: "llama"})
	streamText(t, provider)

	if _, sent := requests[0].headers["Authorization"]; sent {
		t.Error("an empty bearer is rejected by some local runtimes and must not be sent")
	}
}

func TestOpenAIRetriesWithoutReasoningWhenRejected(t *testing.T) {
	var requests []capturedRequest
	server := fakeProvider(t, `data: {"choices":[{"delta":{"content":"ok"}}]}

data: [DONE]

`, &requests, http.StatusBadRequest)

	provider, _ := New(Settings{
		Name: "openai", Kind: KindOpenAI, APIKey: "sk", BaseURL: server.URL,
		Model: "gpt-4o-mini", LowReasoning: true,
	})

	if got := streamText(t, provider); got != "ok" {
		t.Errorf("got %q", got)
	}
	if len(requests) != 2 {
		t.Fatalf("expected a retry, got %d requests", len(requests))
	}
	if requests[0].body["reasoning_effort"] != "low" {
		t.Error("the first attempt should ask for low reasoning")
	}
	if _, present := requests[1].body["reasoning_effort"]; present {
		t.Error("the retry should drop the parameter")
	}
}

func TestOpenRouterUsesItsOwnReasoningShape(t *testing.T) {
	var requests []capturedRequest
	server := fakeProvider(t, "data: [DONE]\n\n", &requests)

	// The style is chosen by host, so a custom provider pointing at OpenRouter
	// gets the same treatment.
	provider, _ := New(Settings{
		Name: "openrouter", Kind: KindOpenRouter, APIKey: "sk", Model: "openai/gpt-4o-mini",
		BaseURL: server.URL + "/openrouter.ai", LowReasoning: true,
	})
	streamText(t, provider)

	reasoning, ok := requests[0].body["reasoning"].(map[string]any)
	if !ok {
		t.Fatalf("expected OpenRouter's reasoning object, got %v", requests[0].body)
	}
	if reasoning["effort"] != "low" || reasoning["exclude"] != true {
		t.Errorf("got %v", reasoning)
	}
	if _, present := requests[0].body["reasoning_effort"]; present {
		t.Error("sending both shapes at once is rejected by OpenRouter")
	}
}

func TestGeminiStreamsParts(t *testing.T) {
	var requests []capturedRequest
	server := fakeProvider(t, `data: {"candidates":[{"content":{"parts":[{"text":"Hola"}]}}]}

data: {"candidates":[{"content":{"parts":[{"text":" mundo"}]}}]}

`, &requests)

	provider, _ := New(Settings{Name: "gemini", Kind: KindGemini, APIKey: "key", BaseURL: server.URL, Model: "gemini-2.5-flash"})

	if got := streamText(t, provider); got != "Hola mundo" {
		t.Errorf("got %q", got)
	}
	if !strings.HasSuffix(requests[0].path, ":streamGenerateContent") {
		t.Errorf("path = %q", requests[0].path)
	}
	if requests[0].query != "alt=sse" {
		t.Errorf("query = %q, Gemini needs alt=sse to answer with events", requests[0].query)
	}
	if requests[0].headers.Get("x-goog-api-key") != "key" {
		t.Error("the key goes in x-goog-api-key")
	}
}

func TestStreamStopsWhenTheCallerCancels(t *testing.T) {
	var requests []capturedRequest
	server := fakeProvider(t, `data: {"choices":[{"delta":{"content":"one"}}]}

data: {"choices":[{"delta":{"content":"two"}}]}

`, &requests)

	provider, _ := New(Settings{Name: "openai", Kind: KindOpenAI, APIKey: "sk", BaseURL: server.URL, Model: "m"})

	var chunks []string
	err := provider.Stream(context.Background(), Request{Prompt: "hi"}, func(chunk string) error {
		chunks = append(chunks, chunk)
		return io.EOF
	})

	if err != io.EOF {
		t.Fatalf("emit's error should end the stream, got %v", err)
	}
	if len(chunks) != 1 {
		t.Errorf("got %d chunks, want 1", len(chunks))
	}
}

func TestValidateReportsWhatIsMissing(t *testing.T) {
	provider, _ := New(Settings{Name: "openai", Kind: KindOpenAI, BaseURL: "https://example.test", Model: "m"})

	if err := provider.Validate(); err == nil || !strings.Contains(err.Error(), "API key") {
		t.Fatalf("got %v", err)
	}
}
