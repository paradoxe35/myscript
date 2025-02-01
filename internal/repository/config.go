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
	return logChange(tx, n, "CREATE")
}

func (n *Config) AfterUpdate(tx *gorm.DB) error {
	return logChange(tx, n, "UPDATE")
}

func (n *Config) AfterDelete(tx *gorm.DB) error {
	return logChange(tx, n, "DELETE")
}

type ConfigRepository struct {
	BaseRepository
}

func NewConfigRepository() *ConfigRepository {
	return &ConfigRepository{}
}

// Functions

func (*ConfigRepository) GetConfig(db *gorm.DB) *Config {
	var config Config
	db.First(&config)
	return &config
}

func (c *ConfigRepository) SaveConfig(db *gorm.DB, config *Config) {
	if newConfig := c.GetConfig(db); newConfig != nil {
		config.ID = newConfig.ID
	}

	db.Save(config)
}
