// Copyright (c) 2024
// Licensed under the MIT License. See LICENSE file in the root directory.

package repository

import (
	"myscript/internal/ai"
	"testing"

	"gorm.io/datatypes"
)

func newProviders(t *testing.T) *AIProviderRepository {
	t.Helper()

	mainDB, unsynced := newStores(t)
	return NewAIProviderRepository(mainDB, unsynced)
}

func TestListStartsWithTheBuiltInProviders(t *testing.T) {
	providers := newProviders(t).List()

	if len(providers) != len(ai.BuiltIn()) {
		t.Fatalf("got %d providers, want %d", len(providers), len(ai.BuiltIn()))
	}

	for i, name := range ai.BuiltIn() {
		provider := providers[i]
		if provider.Name != name {
			t.Errorf("provider %d = %q, want %q", i, provider.Name, name)
		}
		if provider.BaseURL != ai.DefaultBaseURL(name) || provider.Model != ai.DefaultModel(name) {
			t.Errorf("%s should come with defaults, got %+v", name, provider)
		}
		if provider.Custom {
			t.Errorf("%s should not be marked custom", name)
		}
	}
}

func TestSaveOverridesABuiltIn(t *testing.T) {
	repo := newProviders(t)

	if err := repo.Save(AIProvider{Name: ai.KindOpenAI, Model: "gpt-4o", Temperature: 0.2}); err != nil {
		t.Fatal(err)
	}

	provider, ok := repo.Find(ai.KindOpenAI)
	if !ok {
		t.Fatal("openai should still be listed")
	}
	if provider.Model != "gpt-4o" || provider.Temperature != 0.2 {
		t.Errorf("got %+v", provider)
	}
	if provider.BaseURL != ai.DefaultBaseURL(ai.KindOpenAI) {
		t.Errorf("BaseURL = %q, the default should survive", provider.BaseURL)
	}
}

func TestSaveAddsACustomProvider(t *testing.T) {
	repo := newProviders(t)

	err := repo.Save(AIProvider{Name: "local", BaseURL: "http://localhost:1234/v1", Model: "llama", NoAPIKey: true})
	if err != nil {
		t.Fatal(err)
	}

	provider, ok := repo.Find("local")
	if !ok {
		t.Fatal("the provider should be listed")
	}
	if !provider.Custom || provider.Kind != ai.KindOpenAICompatible {
		t.Errorf("got %+v", provider)
	}
}

func TestCustomProvidersNeedABaseURL(t *testing.T) {
	if err := newProviders(t).Save(AIProvider{Name: "local", Model: "llama"}); err == nil {
		t.Fatal("expected an error")
	}
}

func TestCustomProviderNamesAreValidated(t *testing.T) {
	repo := newProviders(t)

	if err := repo.Save(AIProvider{Name: "my provider", BaseURL: "http://x"}); err == nil {
		t.Error("a space should be rejected")
	}
	if err := repo.Save(AIProvider{Name: "  "}); err == nil {
		t.Error("an empty name should be rejected")
	}
}

func TestCustomProviderNamesAreUniqueRegardlessOfCase(t *testing.T) {
	repo := newProviders(t)
	repo.Save(AIProvider{Name: "local", BaseURL: "http://localhost:1234/v1"})

	if err := repo.Save(AIProvider{Name: "LOCAL", BaseURL: "http://localhost:9999/v1"}); err == nil {
		t.Fatal("expected a duplicate to be refused")
	}
}

func TestDeleteRemovesACustomProviderAndItsKey(t *testing.T) {
	mainDB, unsynced := newStores(t)
	repo := NewAIProviderRepository(mainDB, unsynced)

	repo.Save(AIProvider{Name: "local", BaseURL: "http://localhost:1234/v1", Model: "llama"})
	repo.SetAPIKey("local", "secret")
	repo.SetActive("local")

	if err := repo.Delete("local"); err != nil {
		t.Fatal(err)
	}

	if _, ok := repo.Find("local"); ok {
		t.Error("the provider should be gone")
	}
	if key := repo.APIKey("local"); key != "" {
		t.Error("its API key should be gone too")
	}
	if active := repo.Active(); active != ai.KindOpenAI {
		t.Errorf("active = %q, deleting the active provider should fall back", active)
	}
	if chosen := NewDeviceSettingsRepository(unsynced).Get().AIProvider; chosen != "" {
		t.Errorf("the device still names %q", chosen)
	}
}

func TestBuiltInProvidersCannotBeDeleted(t *testing.T) {
	if err := newProviders(t).Delete(ai.KindGemini); err == nil {
		t.Fatal("expected an error")
	}
}

func TestActiveDefaultsToOpenAI(t *testing.T) {
	if active := newProviders(t).Active(); active != ai.KindOpenAI {
		t.Errorf("got %q", active)
	}
}

func TestSetActiveRejectsUnknownProviders(t *testing.T) {
	if err := newProviders(t).SetActive("nowhere"); err == nil {
		t.Fatal("expected an error")
	}
}

func TestSettingsCarryTheStoredKey(t *testing.T) {
	repo := newProviders(t)
	repo.SetAPIKey(ai.KindOpenAI, "sk-test")

	settings, err := repo.Settings(ai.KindOpenAI)
	if err != nil {
		t.Fatal(err)
	}
	if settings.APIKey != "sk-test" {
		t.Errorf("APIKey = %q", settings.APIKey)
	}
	if settings.Temperature == 0 {
		t.Error("Settings should be resolved")
	}
}

func TestConfiguredNeedsEverythingTheProviderAsksFor(t *testing.T) {
	repo := newProviders(t)

	if repo.Configured(ai.KindOpenAI) {
		t.Error("a provider without an API key is not configured")
	}

	repo.SetAPIKey(ai.KindOpenAI, "sk-test")
	if !repo.Configured(ai.KindOpenAI) {
		t.Error("a built-in with a key, a model and a URL is configured")
	}
}

func TestLocalProvidersAreConfiguredWithoutAKey(t *testing.T) {
	repo := newProviders(t)
	repo.Save(AIProvider{Name: "local", BaseURL: "http://localhost:1234/v1", Model: "llama", NoAPIKey: true})

	if !repo.Configured("local") {
		t.Error("a provider that takes no key needs only a URL and a model")
	}
}

func TestProviderSettingsSurviveAReload(t *testing.T) {
	mainDB, unsynced := newStores(t)

	NewAIProviderRepository(mainDB, unsynced).
		Save(AIProvider{Name: "local", BaseURL: "http://localhost:1234/v1", Model: "llama"})

	reloaded, ok := NewAIProviderRepository(mainDB, unsynced).Find("local")
	if !ok {
		t.Fatal("the provider should be read back from the database")
	}
	if reloaded.Model != "llama" {
		t.Errorf("got %+v", reloaded)
	}
}

func TestActiveFallsBackWhenTheProviderIsGone(t *testing.T) {
	mainDB, unsynced := newStores(t)

	// A provider retired between releases.
	NewDeviceSettingsRepository(unsynced).Save(DeviceSettings{AIProvider: "anthropic"})

	if active := NewAIProviderRepository(mainDB, unsynced).Active(); active != ai.KindOpenAI {
		t.Errorf("active = %q, want a provider that still exists", active)
	}
}

func TestSetActiveIsKeptOnThisMachineOnly(t *testing.T) {
	mainDB, unsynced := newStores(t)
	repo := NewAIProviderRepository(mainDB, unsynced)
	repo.SetAPIKey(ai.KindGemini, "key")

	if err := repo.SetActive(ai.KindGemini); err != nil {
		t.Fatal(err)
	}

	if chosen := NewDeviceSettingsRepository(unsynced).Get().AIProvider; chosen != ai.KindGemini {
		t.Errorf("device choice = %q", chosen)
	}
	var configs int64
	mainDB.Model(&Config{}).Count(&configs)
	if configs != 0 {
		t.Error("choosing a provider must not touch the synced config")
	}
	if active := repo.Active(); active != ai.KindGemini {
		t.Errorf("active = %q", active)
	}
}

// The list syncs but the keys do not, so another machine's choice may be
// unusable here; the first provider with a key stands in for it.
func TestActivePrefersAProviderUsableOnThisMachine(t *testing.T) {
	mainDB, unsynced := newStores(t)
	repo := NewAIProviderRepository(mainDB, unsynced)

	repo.SetActive(ai.KindGemini)
	repo.SetAPIKey(ai.KindOpenRouter, "key")

	if active := repo.Active(); active != ai.KindOpenRouter {
		t.Errorf("active = %q, want the provider with a key here", active)
	}

	repo.SetAPIKey(ai.KindGemini, "key")
	if active := repo.Active(); active != ai.KindGemini {
		t.Errorf("active = %q, the choice should stand once it has a key", active)
	}
}

func TestActiveKeepsTheChoiceWhenNothingIsUsable(t *testing.T) {
	repo := newProviders(t)
	repo.SetActive(ai.KindGemini)

	if active := repo.Active(); active != ai.KindGemini {
		t.Errorf("active = %q, the UI should point at the chosen provider to set up", active)
	}
}

func TestAKeylessProviderIsUsableWithoutAKey(t *testing.T) {
	repo := newProviders(t)
	repo.Save(AIProvider{Name: "local", BaseURL: "http://localhost:1234/v1", Model: "llama", NoAPIKey: true})

	if active := repo.Active(); active != "local" {
		t.Errorf("active = %q, want the only usable provider", active)
	}
}

func TestOpenRouterIsBuiltIn(t *testing.T) {
	provider, ok := newProviders(t).Find(ai.KindOpenRouter)

	if !ok {
		t.Fatal("openrouter should be listed")
	}
	if provider.Custom {
		t.Error("a built-in is not custom and cannot be deleted")
	}
	if provider.BaseURL != ai.DefaultBaseURL(ai.KindOpenRouter) {
		t.Errorf("BaseURL = %q", provider.BaseURL)
	}
}

// A JSON null decodes to a nil map; writing to it must not panic.
func TestSaveWhenTheProvidersColumnHoldsNull(t *testing.T) {
	for _, raw := range []string{"null", "", "{}", "not json"} {
		t.Run(raw, func(t *testing.T) {
			mainDB, unsynced := newStores(t)
			NewConfigRepository(mainDB).SaveConfig(&Config{AIProviders: datatypes.JSON(raw)})

			repo := NewAIProviderRepository(mainDB, unsynced)
			if err := repo.Save(AIProvider{Name: "openrouter", Model: "openrouter/free"}); err != nil {
				t.Fatalf("save: %v", err)
			}

			provider, ok := repo.Find("openrouter")
			if !ok || provider.Model != "openrouter/free" {
				t.Errorf("got %+v, ok=%v", provider, ok)
			}
		})
	}
}
