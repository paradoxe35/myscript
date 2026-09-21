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

// A provider as the app sees it: the identity shared through Config merged
// with this machine's model choice. The key lives in the secret store.
type AIProvider struct {
	Name         string
	Kind         string
	BaseURL      string
	Model        string
	Temperature  float64
	NoAPIKey     bool
	LowReasoning bool
	Custom       bool
}

// What Config carries for a provider, so that the synced payload can never
// hold a per-device field.
type storedAIProvider struct {
	Name     string `json:"name"`
	Kind     string `json:"kind"`
	BaseURL  string `json:"base_url,omitempty"`
	NoAPIKey bool   `json:"no_api_key,omitempty"`
	Custom   bool   `json:"custom,omitempty"`
}

func (p storedAIProvider) identity() AIProvider {
	return AIProvider{Name: p.Name, Kind: p.Kind, BaseURL: p.BaseURL, NoAPIKey: p.NoAPIKey, Custom: p.Custom}
}

func (p AIProvider) stored() storedAIProvider {
	return storedAIProvider{Name: p.Name, Kind: p.Kind, BaseURL: p.BaseURL, NoAPIKey: p.NoAPIKey, Custom: p.Custom}
}

func (p AIProvider) deviceSettings() AIProviderSettings {
	return AIProviderSettings{Name: p.Name, Model: p.Model, Temperature: p.Temperature, LowReasoning: p.LowReasoning}
}

func (p AIProvider) with(settings AIProviderSettings) AIProvider {
	p.Model = settings.Model
	p.Temperature = settings.Temperature
	p.LowReasoning = settings.LowReasoning
	return p
}

// The provider list is shared through Config; the active choice, the model
// and the keys are per machine.
type AIProviderRepository struct {
	config         *ConfigRepository
	device         *DeviceSettingsRepository
	deviceSettings *AIProviderSettingsRepository
	secrets        *SecretRepository
}

func NewAIProviderRepository(mainDB, unSyncedDB *gorm.DB) *AIProviderRepository {
	return &AIProviderRepository{
		config:         NewConfigRepository(mainDB),
		device:         NewDeviceSettingsRepository(unSyncedDB),
		deviceSettings: NewAIProviderSettingsRepository(unSyncedDB),
		secrets:        NewSecretRepository(unSyncedDB),
	}
}

func (r *AIProviderRepository) List() []AIProvider {
	stored := r.stored()

	providers := make([]AIProvider, 0, len(stored)+len(ai.BuiltIn()))
	for _, kind := range ai.BuiltIn() {
		provider := stored[kind].identity().with(r.deviceSettings.Get(kind))
		providers = append(providers, withDefaults(provider, kind))
	}

	custom := make([]AIProvider, 0, len(stored))
	for name, provider := range stored {
		if provider.Custom && !ai.IsBuiltIn(name) {
			custom = append(custom, provider.identity().with(r.deviceSettings.Get(name)))
		}
	}
	sort.Slice(custom, func(i, j int) bool {
		return strings.ToLower(custom[i].Name) < strings.ToLower(custom[j].Name)
	})

	return append(providers, custom...)
}

func (r *AIProviderRepository) Find(name string) (AIProvider, bool) {
	return find(r.List(), name)
}

func find(providers []AIProvider, name string) (AIProvider, bool) {
	for _, provider := range providers {
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

	stored[provider.Name] = provider.stored()
	if err := r.write(stored); err != nil {
		return err
	}
	return r.deviceSettings.Save(provider.deviceSettings())
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
	if err := r.deviceSettings.Delete(name); err != nil {
		return err
	}
	if err := r.secrets.Delete(AIProviderSecret(name)); err != nil {
		return err
	}

	if settings := r.device.Get(); strings.EqualFold(settings.AIProvider, name) {
		settings.AIProvider = ""
		r.device.Save(settings)
	}
	return nil
}

// The provider to use: this machine's choice when it is usable here (a key
// is stored, or none is needed), otherwise the first usable one in the list.
// With nothing usable the choice stands, so the UI can point at what to set up.
func (r *AIProviderRepository) Active() string {
	providers := r.List()
	chosen, ok := find(providers, r.device.Get().AIProvider)
	if !ok {
		chosen = providers[0]
	}

	if r.configured(chosen) {
		return chosen.Name
	}
	for _, provider := range providers {
		if r.configured(provider) {
			return provider.Name
		}
	}
	return chosen.Name
}

func (r *AIProviderRepository) SetActive(name string) error {
	provider, ok := r.Find(name)
	if !ok {
		return fmt.Errorf("no provider named %q", name)
	}

	settings := r.device.Get()
	settings.AIProvider = provider.Name
	r.device.Save(settings)
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
	return r.resolve(provider), nil
}

func (r *AIProviderRepository) resolve(provider AIProvider) ai.Settings {
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
	}.Resolved()
}

func (r *AIProviderRepository) Configured(name string) bool {
	provider, ok := r.Find(name)
	return ok && r.configured(provider)
}

func (r *AIProviderRepository) configured(provider AIProvider) bool {
	client, err := ai.New(r.resolve(provider))
	if err != nil {
		return false
	}
	return client.Validate() == nil
}

func (r *AIProviderRepository) stored() map[string]storedAIProvider {
	providers := map[string]storedAIProvider{}

	if raw := r.config.GetConfig().AIProviders; len(raw) > 0 {
		// A JSON null decodes to a nil map, and writing to it would panic.
		if err := json.Unmarshal(raw, &providers); err != nil || providers == nil {
			providers = map[string]storedAIProvider{}
		}
	}

	for name, provider := range providers {
		provider.Name = name
		providers[name] = provider
	}
	return providers
}

func (r *AIProviderRepository) write(providers map[string]storedAIProvider) error {
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
