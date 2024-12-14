package database

import (
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

const DB_NAME = "database.sqlite"

func CreateDatabase(homeDir string) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(homeDir+"/"+DB_NAME), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	return db
}
