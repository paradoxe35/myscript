package database

import (
	"myscript/internal/repository"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

const DB_NAME = "data.db"

func NewDatabase(homeDir string) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(homeDir+"/"+DB_NAME), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	// Migrate schemas
	db.AutoMigrate(&repository.Config{})
	db.AutoMigrate(&repository.Page{})
	db.AutoMigrate(&repository.Cache{})

	return db
}
