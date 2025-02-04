package repository

import "gorm.io/gorm"

// !SYNCED MODEL

type Config struct {
	gorm.Model
	NotionApiKey *string `gorm:"column:notion_api_key"`
	OpenAIApiKey *string `gorm:"column:openai_api_key"`
	GroqApiKey   *string `gorm:"column:groq_api_key"`

	TranscriberSource string  `gorm:"column:transcriber_source;default:local"` // local, openai, witai, groq
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
