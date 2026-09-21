// Copyright (c) 2024
// Licensed under the MIT License. See LICENSE file in the root directory.

package main

import (
	"context"
	"fmt"
	"log/slog"
	"myscript/internal/ai"
	"myscript/internal/repository"
	"runtime/debug"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

const (
	eventAIChunk = "on-ai-completion-chunk"
	eventAIDone  = "on-ai-completion-done"
	eventAIError = "on-ai-completion-error"

	aiTestTimeout = 30 * time.Second
)

type AIProvider struct {
	Name         string
	Kind         string
	BaseURL      string
	Model        string
	Temperature  float64
	NoAPIKey     bool
	LowReasoning bool
	Custom       bool
	HasAPIKey    bool
	Configured   bool
	Active       bool
}

type AICompletionRequest struct {
	Task     string
	Text     string
	Command  string
	Provider string
}

type AICompletionEvent struct {
	ID    string
	Chunk string
	Error string
}

// Wails logs a panic in a binding but never settles the promise, so the
// window would wait forever; return it as an error instead.
func recoverAsError(err *error) {
	if panicked := recover(); panicked != nil {
		slog.Error("Recovered from a panic in a bound method",
			"panic", panicked, "stack", string(debug.Stack()))
		*err = fmt.Errorf("unexpected failure: %v", panicked)
	}
}

func (a *App) aiProviders() *repository.AIProviderRepository {
	return repository.NewAIProviderRepository(a.mainDB, a.unSyncedDB)
}

func (a *App) GetAIProviders() []AIProvider {
	repo := a.aiProviders()
	active := repo.Active()

	stored := repo.List()
	providers := make([]AIProvider, 0, len(stored))

	for _, provider := range stored {
		providers = append(providers, AIProvider{
			Name:         provider.Name,
			Kind:         provider.Kind,
			BaseURL:      provider.BaseURL,
			Model:        provider.Model,
			Temperature:  provider.Temperature,
			NoAPIKey:     provider.NoAPIKey,
			LowReasoning: provider.LowReasoning,
			Custom:       provider.Custom,
			HasAPIKey:    repo.APIKey(provider.Name) != "",
			Configured:   repo.Configured(provider.Name),
			Active:       strings.EqualFold(provider.Name, active),
		})
	}

	return providers
}

func (a *App) GetActiveAIProvider() string {
	return a.aiProviders().Active()
}

func (a *App) SetActiveAIProvider(name string) (err error) {
	defer recoverAsError(&err)
	return a.aiProviders().SetActive(name)
}

func (a *App) SaveAIProvider(provider AIProvider, apiKey string) (err error) {
	defer recoverAsError(&err)

	repo := a.aiProviders()

	record := repository.AIProvider{
		Name:         provider.Name,
		Kind:         provider.Kind,
		BaseURL:      strings.TrimSpace(provider.BaseURL),
		Model:        strings.TrimSpace(provider.Model),
		Temperature:  provider.Temperature,
		NoAPIKey:     provider.NoAPIKey,
		LowReasoning: provider.LowReasoning,
		Custom:       provider.Custom,
	}

	if err := repo.Save(record); err != nil {
		return err
	}

	return repo.SetAPIKey(provider.Name, apiKey)
}

func (a *App) DeleteAIProvider(name string) (err error) {
	defer recoverAsError(&err)
	return a.aiProviders().Delete(name)
}

func (a *App) GetAIProviderAPIKey(name string) string {
	return a.aiProviders().APIKey(name)
}

// Uses the values on screen so an unsaved edit can be tried before saving.
func (a *App) ListAIModels(provider AIProvider, apiKey string) (models []ai.ModelInfo, err error) {
	defer recoverAsError(&err)

	settings := ai.Settings{
		Name:    provider.Name,
		Kind:    provider.Kind,
		APIKey:  apiKey,
		BaseURL: provider.BaseURL,
		Custom:  provider.Custom,
	}

	if apiKey == "" {
		settings.APIKey = a.aiProviders().APIKey(provider.Name)
	}

	return ai.ListModels(context.Background(), settings)
}

func (a *App) TestAIProvider(provider AIProvider, apiKey string) (err error) {
	defer recoverAsError(&err)

	settings := ai.Settings{
		Name:         provider.Name,
		Kind:         provider.Kind,
		APIKey:       apiKey,
		BaseURL:      provider.BaseURL,
		Model:        provider.Model,
		Temperature:  provider.Temperature,
		NoAPIKey:     provider.NoAPIKey,
		LowReasoning: provider.LowReasoning,
		Custom:       provider.Custom,
	}

	if apiKey == "" {
		settings.APIKey = a.aiProviders().APIKey(provider.Name)
	}

	client, err := ai.New(settings)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), aiTestTimeout)
	defer cancel()

	_, err = ai.Complete(ctx, client, ai.Request{
		System: "Reply with the single word: ready.",
		Prompt: "ready",
	})
	return err
}

// Returns once accepted; chunks arrive as events keyed by the returned id.
func (a *App) StartAICompletion(request AICompletionRequest) (id string, err error) {
	defer recoverAsError(&err)

	repo := a.aiProviders()

	name := request.Provider
	if name == "" {
		name = repo.Active()
	}

	settings, err := repo.Settings(name)
	if err != nil {
		return "", err
	}

	client, err := ai.New(settings)
	if err != nil {
		return "", err
	}
	if err := client.Validate(); err != nil {
		return "", err
	}

	prompt, err := ai.BuildRequest(ai.Task(request.Task), request.Text, request.Command)
	if err != nil {
		return "", err
	}

	id = uuid.New().String()
	ctx, cancel := context.WithCancel(context.Background())
	a.aiCompletions.start(id, cancel)

	go func() {
		defer a.aiCompletions.finish(id)

		err := client.Stream(ctx, prompt, func(chunk string) error {
			runtime.EventsEmit(a.ctx, eventAIChunk, AICompletionEvent{ID: id, Chunk: chunk})
			return nil
		})

		switch {
		case ctx.Err() != nil:
			runtime.EventsEmit(a.ctx, eventAIDone, AICompletionEvent{ID: id})
		case err != nil:
			slog.Error("AI completion failed", "provider", name, "error", err)
			runtime.EventsEmit(a.ctx, eventAIError, AICompletionEvent{ID: id, Error: err.Error()})
		default:
			runtime.EventsEmit(a.ctx, eventAIDone, AICompletionEvent{ID: id})
		}
	}()

	return id, nil
}

func (a *App) CancelAICompletion(id string) {
	a.aiCompletions.cancel(id)
}

func (a *App) HasConfiguredAIProvider() bool {
	repo := a.aiProviders()
	return repo.Configured(repo.Active())
}

// aiCompletions tracks the streams in flight so the editor can abandon one.
type aiCompletions struct {
	mu      sync.Mutex
	cancels map[string]context.CancelFunc
}

func newAICompletions() *aiCompletions {
	return &aiCompletions{cancels: map[string]context.CancelFunc{}}
}

func (c *aiCompletions) start(id string, cancel context.CancelFunc) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.cancels[id] = cancel
}

func (c *aiCompletions) finish(id string) {
	c.mu.Lock()
	cancel := c.cancels[id]
	delete(c.cancels, id)
	c.mu.Unlock()

	if cancel != nil {
		cancel()
	}
}

func (c *aiCompletions) cancel(id string) {
	c.mu.Lock()
	cancel := c.cancels[id]
	c.mu.Unlock()

	if cancel != nil {
		cancel()
	}
}

func (c *aiCompletions) cancelAll() {
	c.mu.Lock()
	cancels := c.cancels
	c.cancels = map[string]context.CancelFunc{}
	c.mu.Unlock()

	for _, cancel := range cancels {
		cancel()
	}
}
