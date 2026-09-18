// Copyright (c) 2024
// Licensed under the MIT License. See LICENSE file in the root directory.

package ai

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestDecodeModelsReadsBothListShapes(t *testing.T) {
	body := `{
	  "data": [{"id":"gpt-4o","display_name":"GPT-4o"},{"id":""}],
	  "models": [
	    {"name":"models/gemini-2.5-flash","displayName":"Gemini Flash","supportedGenerationMethods":["generateContent"]},
	    {"name":"models/embedding-001","displayName":"Embedding","supportedGenerationMethods":["embedContent"]}
	  ]
	}`

	models, err := decodeModels(strings.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}

	ids := make([]string, 0, len(models))
	for _, model := range models {
		ids = append(ids, model.ID)
	}

	want := []string{"gpt-4o", "gemini-2.5-flash"}
	if len(ids) != len(want) {
		t.Fatalf("got %v, want %v", ids, want)
	}
	for i, id := range want {
		if ids[i] != id {
			t.Errorf("model %d = %q, want %q", i, ids[i], id)
		}
	}
}

func TestListModelsUsesTheProviderEndpoint(t *testing.T) {
	var path, query, key string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path, query, key = r.URL.Path, r.URL.RawQuery, r.Header.Get("x-api-key")
		io.WriteString(w, `{"data":[{"id":"claude-3-5-haiku-latest"}]}`)
	}))
	defer server.Close()

	models, err := ListModels(context.Background(), Settings{
		Name: "anthropic", Kind: KindAnthropic, APIKey: "key", BaseURL: server.URL,
	})
	if err != nil {
		t.Fatal(err)
	}

	if len(models) != 1 || models[0].ID != "claude-3-5-haiku-latest" {
		t.Errorf("got %+v", models)
	}
	if path != "/v1/models" || query != "limit=1000" {
		t.Errorf("requested %s?%s", path, query)
	}
	if key != "key" {
		t.Errorf("x-api-key = %q", key)
	}
}

func TestListModelsReportsProviderErrors(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer server.Close()

	_, err := ListModels(context.Background(), Settings{Name: "openai", Kind: KindOpenAI, BaseURL: server.URL})
	if err == nil || !strings.Contains(err.Error(), "API key") {
		t.Fatalf("got %v", err)
	}
}
