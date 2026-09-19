// Copyright (c) 2024
// Licensed under the MIT License. See LICENSE file in the root directory.

package repository

import (
	"log/slog"
	"myscript/internal/secrets"
	"time"

	"gorm.io/gorm"
)

// UNSYNCED MODEL
//
// Credentials live here rather than in Config so they are never uploaded to
// Drive, and they are encrypted so the file alone gives nothing away.
type Secret struct {
	Name      string `gorm:"primaryKey"`
	Value     string
	CreatedAt time.Time
	UpdatedAt time.Time
}

const (
	SecretNotionAPIKey = "notion.api_key"

	// Kept only so a key written before hosted services were unified can be adopted.
	SecretSpeechOpenAIAPIKey = "speech.openai.api_key"
	SecretSpeechGroqAPIKey   = "speech.groq.api_key"
)

// SpeechServiceSecret keys the credential per service, so switching between
// them does not throw away the key you are not using.
func SpeechServiceSecret(preset string) string {
	return "speech.remote." + preset + ".api_key"
}

func AIProviderSecret(provider string) string {
	return "ai." + provider + ".api_key"
}

type SecretRepository struct {
	BaseRepository
}

func NewSecretRepository(unSyncedDB *gorm.DB) *SecretRepository {
	return &SecretRepository{BaseRepository: BaseRepository{db: unSyncedDB}}
}

func (r *SecretRepository) Get(name string) string {
	var secret Secret
	if err := r.db.Where("name = ?", name).First(&secret).Error; err != nil {
		return ""
	}

	value, err := secrets.Decrypt(secret.Value)
	if err != nil {
		slog.Warn("Stored secret could not be decrypted", "name", name, "error", err)
		return ""
	}

	return value
}

func (r *SecretRepository) Set(name, value string) error {
	if value == "" {
		return r.Delete(name)
	}

	encrypted, err := secrets.Encrypt(value)
	if err != nil {
		return err
	}

	return r.db.Save(&Secret{Name: name, Value: encrypted}).Error
}

func (r *SecretRepository) Delete(name string) error {
	return r.db.Where("name = ?", name).Delete(&Secret{}).Error
}

func (r *SecretRepository) Has(name string) bool {
	var count int64
	r.db.Model(&Secret{}).Where("name = ?", name).Count(&count)
	return count > 0
}
