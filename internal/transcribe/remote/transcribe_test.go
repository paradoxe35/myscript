// Copyright (c) 2024
// Licensed under the MIT License. See LICENSE file in the root directory.

package remote

import (
	"context"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type capturedRequest struct {
	path    string
	headers http.Header
	fields  map[string]string
	audio   []byte
	body    string
}

func fakeService(t *testing.T, reply string, captured *[]capturedRequest, status ...int) *httptest.Server {
	t.Helper()

	call := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		record := capturedRequest{path: r.URL.Path, headers: r.Header.Clone(), fields: map[string]string{}}

		contentType := r.Header.Get("Content-Type")
		if mediaType, params, err := mime.ParseMediaType(contentType); err == nil && strings.HasPrefix(mediaType, "multipart/") {
			reader := multipart.NewReader(r.Body, params["boundary"])
			for {
				part, err := reader.NextPart()
				if err != nil {
					break
				}
				value, _ := io.ReadAll(part)
				if part.FormName() == "file" {
					record.audio = value
				} else {
					record.fields[part.FormName()] = string(value)
				}
			}
		} else {
			raw, _ := io.ReadAll(r.Body)
			record.body = string(raw)
		}

		*captured = append(*captured, record)

		if call < len(status) && status[call] != http.StatusOK {
			code := status[call]
			call++
			w.WriteHeader(code)
			io.WriteString(w, `{"error":{"message":"unsupported field"}}`)
			return
		}
		call++
		io.WriteString(w, reply)
	}))

	t.Cleanup(server.Close)
	return server
}

func TestOpenAIProtocolPostsMultipart(t *testing.T) {
	var requests []capturedRequest
	server := fakeService(t, `{"text":"hello there"}`, &requests)

	text, err := Transcribe(context.Background(),
		Settings{Preset: CustomPreset, BaseURL: server.URL, Model: "whisper-1", APIKey: "sk-test"},
		[]byte("RIFFfake"), "fr")
	if err != nil {
		t.Fatal(err)
	}

	if text != "hello there" {
		t.Errorf("text = %q", text)
	}
	if requests[0].path != "/audio/transcriptions" {
		t.Errorf("path = %q", requests[0].path)
	}
	if requests[0].headers.Get("Authorization") != "Bearer sk-test" {
		t.Errorf("Authorization = %q", requests[0].headers.Get("Authorization"))
	}
	if requests[0].fields["model"] != "whisper-1" {
		t.Errorf("model = %q", requests[0].fields["model"])
	}
	if requests[0].fields["language"] != "fr" {
		t.Errorf("language = %q", requests[0].fields["language"])
	}
	if string(requests[0].audio) != "RIFFfake" {
		t.Errorf("the recording should be posted verbatim, got %q", requests[0].audio)
	}
}

func TestGptModelsUseTheRepeatedLanguageField(t *testing.T) {
	var requests []capturedRequest
	server := fakeService(t, `{"text":"ok"}`, &requests)

	Transcribe(context.Background(),
		Settings{Preset: CustomPreset, BaseURL: server.URL, Model: "gpt-4o-mini-transcribe"},
		[]byte("wav"), "de")

	if _, ok := requests[0].fields["languages[]"]; !ok {
		t.Errorf("gpt- models take languages[], got %v", requests[0].fields)
	}
}

func TestLanguageIsOmittedWhenUnset(t *testing.T) {
	var requests []capturedRequest
	server := fakeService(t, `{"text":"ok"}`, &requests)

	Transcribe(context.Background(),
		Settings{Preset: CustomPreset, BaseURL: server.URL, Model: "whisper-1"}, []byte("wav"), "")

	if _, present := requests[0].fields["language"]; present {
		t.Error("an unset language should let the service detect it")
	}
}

func TestNoAuthorizationWithoutAKey(t *testing.T) {
	var requests []capturedRequest
	server := fakeService(t, `{"text":"ok"}`, &requests)

	Transcribe(context.Background(),
		Settings{Preset: CustomPreset, BaseURL: server.URL, Model: "whisper-1"}, []byte("wav"), "")

	if _, sent := requests[0].headers["Authorization"]; sent {
		t.Error("a local runtime takes no key and some reject an empty bearer")
	}
}

func TestServiceErrorsCarryTheMessage(t *testing.T) {
	var requests []capturedRequest
	server := fakeService(t, "", &requests, http.StatusUnauthorized)

	_, err := Transcribe(context.Background(),
		Settings{Preset: CustomPreset, BaseURL: server.URL, Model: "whisper-1", APIKey: "bad"},
		[]byte("wav"), "")

	if err == nil || !strings.Contains(err.Error(), "unsupported field") {
		t.Fatalf("got %v", err)
	}
	if status, ok := err.(*StatusError); !ok || status.Status != http.StatusUnauthorized {
		t.Errorf("the status should survive, got %#v", err)
	}
}

func TestGeminiProtocolSendsInlineAudio(t *testing.T) {
	var requests []capturedRequest
	server := fakeService(t, `{"candidates":[{"content":{"parts":[{"text":" bonjour "}]}}]}`, &requests)

	text, err := Transcribe(context.Background(),
		Settings{Preset: "gemini", BaseURL: server.URL, Model: "gemini-2.5-flash", APIKey: "key"},
		[]byte("wav"), "fr")
	if err != nil {
		t.Fatal(err)
	}

	if text != "bonjour" {
		t.Errorf("text = %q, it should be trimmed", text)
	}
	if requests[0].path != "/v1beta/models/gemini-2.5-flash:generateContent" {
		t.Errorf("path = %q", requests[0].path)
	}
	if requests[0].headers.Get("x-goog-api-key") != "key" {
		t.Error("the key goes in x-goog-api-key")
	}
	if !strings.Contains(requests[0].body, "inline_data") {
		t.Error("the recording should be inline")
	}
	if !strings.Contains(requests[0].body, "French") {
		t.Error("the language should be named in the prompt")
	}
}

func TestGeminiRetriesWithoutThinkingWhenRejected(t *testing.T) {
	var requests []capturedRequest
	server := fakeService(t, `{"candidates":[{"content":{"parts":[{"text":"ok"}]}}]}`,
		&requests, http.StatusBadRequest)

	text, err := Transcribe(context.Background(),
		Settings{Preset: "gemini", BaseURL: server.URL, Model: "gemini-legacy", APIKey: "key"},
		[]byte("wav"), "")
	if err != nil {
		t.Fatal(err)
	}

	if text != "ok" {
		t.Errorf("text = %q", text)
	}
	if len(requests) != 2 {
		t.Fatalf("expected a retry, got %d requests", len(requests))
	}
	if !strings.Contains(requests[0].body, "thinkingConfig") {
		t.Error("the first attempt should ask the model not to think")
	}
	if strings.Contains(requests[1].body, "thinkingConfig") {
		t.Error("the retry should drop the field")
	}
}

func TestGeminiReportsARefusal(t *testing.T) {
	var requests []capturedRequest
	server := fakeService(t, `{"promptFeedback":{"blockReason":"SAFETY"}}`, &requests)

	_, err := Transcribe(context.Background(),
		Settings{Preset: "gemini", BaseURL: server.URL, Model: "gemini-2.5-flash", APIKey: "key"},
		[]byte("wav"), "")

	if err == nil || !strings.Contains(err.Error(), "SAFETY") {
		t.Fatalf("got %v", err)
	}
}
