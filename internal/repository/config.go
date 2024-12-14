package repository

import "gorm.io/gorm"

type Config struct {
	gorm.Model
	NotionApiKey *string `gorm:"column:notion_api_key"`
}

func GetConfig(db *gorm.DB) *Config {
	var config Config
	db.First(&config)
	return &config
}

func SaveConfig(db *gorm.DB, config *Config) {
	if newConfig := GetConfig(db); newConfig != nil {
		config.ID = newConfig.ID
	}

	db.Save(config)
}
