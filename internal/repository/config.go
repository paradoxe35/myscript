// Copyright (c) 2024
// Licensed under the MIT License. See LICENSE file in the root directory.

package repository

import (
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// !SYNCED MODEL

type Config struct {
	gorm.Model

	TranscriberSource string  `gorm:"column:transcriber_source;default:local"` // local, remote, witai
	SpeechModelID     *string `gorm:"column:speech_model_id"`                  // catalogue id of the local model

	// The hosted transcription service, when TranscriberSource is "remote".
	RemoteProvider string `gorm:"column:remote_provider"`
	RemoteModel    string `gorm:"column:remote_model"`
	RemoteBaseURL  string `gorm:"column:remote_base_url"`

	AIProvider  string         `gorm:"column:ai_provider"`
	AIProviders datatypes.JSON `gorm:"column:ai_providers"`

	// Credentials now live in the secret store, which is never synced. These
	// columns are only read once, to adopt a key written by an older build.
	NotionApiKey *string `gorm:"column:notion_api_key"`
	OpenAIApiKey *string `gorm:"column:openai_api_key"`
	GroqApiKey   *string `gorm:"column:groq_api_key"`

	// Unused since the GGUF catalogue replaced the ggml models; kept so a synced
	// row from an older build still applies.
	LocalWhisperModel *string `gorm:"column:local_whisper_model"`
	LocalWhisperGPU   *bool   `gorm:"column:local_whisper_gpu"`
}

// Hooks
func (n *Config) AfterCreate(tx *gorm.DB) error {
	return logChange(tx, n, OPERATION_SAVE)
}

func (n *Config) AfterUpdate(tx *gorm.DB) error {
	return logChange(tx, n, OPERATION_SAVE)
}

func (n *Config) AfterDelete(tx *gorm.DB) error {
	return logChange(tx, n, OPERATION_DELETE)
}

type ConfigRepository struct {
	BaseRepository
}

func NewConfigRepository(db *gorm.DB) *ConfigRepository {
	return &ConfigRepository{
		BaseRepository: BaseRepository{db: db},
	}
}

// Functions

func (r *ConfigRepository) GetConfig() *Config {
	var config Config
	r.db.First(&config)
	return &config
}

func (r *ConfigRepository) SaveConfig(config *Config) {
	if newConfig := r.GetConfig(); newConfig != nil {
		config.ID = newConfig.ID
	}

	r.db.Save(config)
}
