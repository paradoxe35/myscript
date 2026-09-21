// Copyright (c) 2024
// Licensed under the MIT License. See LICENSE file in the root directory.

package repository

import (
	"encoding/json"
	"fmt"
	"myscript/internal/ai"
	"sort"
	"strings"
	"unicode"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// Credentials live in the secret store so they stay on this machine.
type AIProvider struct {
	Name         string  `json:"name"`
	Kind         string  `json:"kind"`
	BaseURL      string  `json:"base_url,omitempty"`
	Model        string  `json:"model,omitempty"`
	Temperature  float64 `json:"temperature,omitempty"`
	NoAPIKey     bool    `json:"no_api_key,omitempty"`
	LowReasoning bool    `json:"low_reasoning,omitempty"`
	Custom       bool    `json:"custom,omitempty"`
}

type AIProviderRepository struct {
	config  *ConfigRepository
	secrets *SecretRepository
}

func NewAIProviderRepository(mainDB, unSyncedDB *gorm.DB) *AIProviderRepository {
	return &AIProviderRepository{
		config:  NewConfigRepository(mainDB),
		secrets: NewSecretRepository(unSyncedDB),
	}
}

func (r *AIProviderRepository) List() []AIProvider {
	stored := r.stored()

	providers := make([]AIProvider, 0, len(stored)+len(ai.BuiltIn()))
	for _, kind := range ai.BuiltIn() {
		providers = append(providers, withDefaults(stored[kind], kind))
	}

	custom := make([]AIProvider, 0, len(stored))
	for name, provider := range stored {
		if provider.Custom && !ai.IsBuiltIn(name) {
			custom = append(custom, provider)
		}
	}
	sort.Slice(custom, func(i, j int) bool {
		return strings.ToLower(custom[i].Name) < strings.ToLower(custom[j].Name)
	})

	return append(providers, custom...)
}

func (r *AIProviderRepository) Find(name string) (AIProvider, bool) {
	for _, provider := range r.List() {
		if strings.EqualFold(provider.Name, name) {
			return provider, true
		}
	}
	return AIProvider{}, false
}

func (r *AIProviderRepository) Save(provider AIProvider) error {
	provider.Name = strings.TrimSpace(provider.Name)
	if provider.Name == "" {
		return fmt.Errorf("a provider name is required")
	}

	if ai.IsBuiltIn(provider.Name) {
		provider.Custom = false
		provider.Kind = provider.Name
	} else {
		if err := validateCustomName(provider.Name); err != nil {
			return err
		}
		provider.Custom = true
		provider.Kind = ai.KindOpenAICompatible

		if strings.TrimSpace(provider.BaseURL) == "" {
			return fmt.Errorf("a base URL is required for %q", provider.Name)
		}
	}

	stored := r.stored()
	for name := range stored {
		if name != provider.Name && strings.EqualFold(name, provider.Name) {
			return fmt.Errorf("a provider named %q already exists", name)
		}
	}

	stored[provider.Name] = provider
	return r.write(stored)
}

func (r *AIProviderRepository) Delete(name string) error {
	if ai.IsBuiltIn(name) {
		return fmt.Errorf("%s is built in and cannot be removed", name)
	}

	stored := r.stored()
	if _, ok := stored[name]; !ok {
		return fmt.Errorf("no provider named %q", name)
	}
	delete(stored, name)

	if err := r.write(stored); err != nil {
		return err
	}
	if err := r.secrets.Delete(AIProviderSecret(name)); err != nil {
		return err
	}

	if strings.EqualFold(r.Active(), name) {
		return r.SetActive(ai.KindOpenAI)
	}
	return nil
}

// Falls back when the stored choice no longer exists.
func (r *AIProviderRepository) Active() string {
	name := r.config.GetConfig().AIProvider
	if name == "" {
		return ai.KindOpenAI
	}

	for _, provider := range r.List() {
		if strings.EqualFold(provider.Name, name) {
			return provider.Name
		}
	}
	return ai.KindOpenAI
}

func (r *AIProviderRepository) SetActive(name string) error {
	if _, ok := r.Find(name); !ok {
		return fmt.Errorf("no provider named %q", name)
	}

	config := r.config.GetConfig()
	config.AIProvider = name
	r.config.SaveConfig(config)
	return nil
}

func (r *AIProviderRepository) APIKey(name string) string {
	return r.secrets.Get(AIProviderSecret(name))
}

func (r *AIProviderRepository) SetAPIKey(name, key string) error {
	return r.secrets.Set(AIProviderSecret(name), strings.TrimSpace(key))
}

func (r *AIProviderRepository) Settings(name string) (ai.Settings, error) {
	provider, ok := r.Find(name)
	if !ok {
		return ai.Settings{}, fmt.Errorf("no provider named %q", name)
	}

	return ai.Settings{
		Name:         provider.Name,
		Kind:         provider.Kind,
		APIKey:       r.APIKey(provider.Name),
		BaseURL:      provider.BaseURL,
		Model:        provider.Model,
		Temperature:  provider.Temperature,
		NoAPIKey:     provider.NoAPIKey,
		LowReasoning: provider.LowReasoning,
		Custom:       provider.Custom,
	}.Resolved(), nil
}

func (r *AIProviderRepository) Configured(name string) bool {
	settings, err := r.Settings(name)
	if err != nil {
		return false
	}

	provider, err := ai.New(settings)
	if err != nil {
		return false
	}

	return provider.Validate() == nil
}

func (r *AIProviderRepository) stored() map[string]AIProvider {
	providers := map[string]AIProvider{}

	if raw := r.config.GetConfig().AIProviders; len(raw) > 0 {
		// A JSON null decodes to a nil map, and writing to it would panic.
		if err := json.Unmarshal(raw, &providers); err != nil || providers == nil {
			providers = map[string]AIProvider{}
		}
	}

	for name, provider := range providers {
		provider.Name = name
		providers[name] = provider
	}
	return providers
}

func (r *AIProviderRepository) write(providers map[string]AIProvider) error {
	raw, err := json.Marshal(providers)
	if err != nil {
		return err
	}

	config := r.config.GetConfig()
	config.AIProviders = datatypes.JSON(raw)
	r.config.SaveConfig(config)
	return nil
}

func withDefaults(provider AIProvider, kind string) AIProvider {
	provider.Name = kind
	provider.Kind = kind
	provider.Custom = false

	if provider.BaseURL == "" {
		provider.BaseURL = ai.DefaultBaseURL(kind)
	}
	if provider.Model == "" {
		provider.Model = ai.DefaultModel(kind)
	}
	return provider
}

func validateCustomName(name string) error {
	for _, r := range name {
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) && r != '-' && r != '_' {
			return fmt.Errorf("a provider name may only contain letters, numbers, hyphens and underscores")
		}
	}
	return nil
}
