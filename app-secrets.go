// Copyright (c) 2024
// Licensed under the MIT License. See LICENSE file in the root directory.

package main

import (
	"fmt"
	"myscript/internal/repository"
)

// Credentials the settings screen owns, mapped to the names they are stored
// under. Anything outside this list is refused.
var appSecrets = map[string]string{
	"notion": repository.SecretNotionAPIKey,
}

func (a *App) secrets() *repository.SecretRepository {
	return repository.NewSecretRepository(a.unSyncedDB)
}

func (a *App) secretName(key string) (string, error) {
	name, ok := appSecrets[key]
	if !ok {
		return "", fmt.Errorf("unknown secret %q", key)
	}
	return name, nil
}

func (a *App) GetSecret(key string) (string, error) {
	name, err := a.secretName(key)
	if err != nil {
		return "", err
	}
	return a.secrets().Get(name), nil
}

func (a *App) SaveSecret(key, value string) error {
	name, err := a.secretName(key)
	if err != nil {
		return err
	}
	return a.secrets().Set(name, value)
}

func (a *App) notionAPIKey() string {
	return a.secrets().Get(repository.SecretNotionAPIKey)
}
