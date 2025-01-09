package database

import (
	"myscript/internal/repository"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

const DB_NAME = "data.db"

func setupDatabase(homeDir string) (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open(homeDir+"/"+DB_NAME), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	// Enable WAL mode
	if _, err := sqlDB.Exec("PRAGMA journal_mode=WAL;"); err != nil {
		return nil, err
	}

	// Optional: Configure other PRAGMA settings for better performance
	pragmas := []string{
		"PRAGMA busy_timeout=5000;",  // Wait up to 5 seconds when database is locked
		"PRAGMA synchronous=NORMAL;", // Balance between safety and speed
		"PRAGMA cache_size=-2000;",   // Use 2MB of memory for page cache
		"PRAGMA foreign_keys=ON;",    // Enable foreign key constraints
		"PRAGMA temp_store=MEMORY;",  // Store temp tables in memory
	}

	for _, pragma := range pragmas {
		if _, err := sqlDB.Exec(pragma); err != nil {
			return nil, err
		}
	}

	return db, nil
}

func NewDatabase(homeDir string) *gorm.DB {
	db, err := setupDatabase(homeDir)
	if err != nil {
		panic("failed to connect database: " + err.Error())
	}

	// Migrate schemas
	db.AutoMigrate(&repository.Config{})
	db.AutoMigrate(&repository.Page{})
	db.AutoMigrate(&repository.Cache{})

	return db
}
