package database

import (
	"fmt"
	"log"
	"myscript/internal/repository"
	"reflect"

	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

type DatabaseSynchronizer struct {
	sourceDB *gorm.DB
	targetDB *gorm.DB
}

func NewDatabaseSynchronizer(sourceDB *gorm.DB, targetDB *gorm.DB) *DatabaseSynchronizer {
	return &DatabaseSynchronizer{
		sourceDB: sourceDB,
		targetDB: targetDB,
	}
}

func (s *DatabaseSynchronizer) SynchronizeAll() error {
	entities := []interface{}{
		&repository.Config{},
		&repository.Page{},
		&repository.Cache{},
	}

	for _, entity := range entities {
		if err := s.SynchronizeEntity(entity); err != nil {
			return fmt.Errorf("synchronization failed for %T: %v", entity, err)
		}
	}
	return nil
}

func (s *DatabaseSynchronizer) SynchronizeEntity(entity interface{}) error {
	// 1. Schema compatibility check
	if compatible, err := s.areSchemasCompatible(entity); !compatible || err != nil {
		if err != nil {
			return fmt.Errorf("schema check error: %v", err)
		}
		return fmt.Errorf("schemas are incompatible for %T", entity)
	}

	// 2. Synchronization logic
	return s.targetDB.Transaction(func(tx *gorm.DB) error {
		// Get all records from source DB
		records, err := s.getSourceRecords(entity)
		if err != nil {
			return err
		}

		// Upsert records in target DB
		for _, record := range records {
			result := tx.Save(record)
			if result.Error != nil {
				log.Printf("Failed to sync record %v: %v", record, result.Error)
				continue
			}
		}
		return nil
	})
}

func (s *DatabaseSynchronizer) areSchemasCompatible(entity interface{}) (bool, error) {
	sourceMigrator := s.sourceDB.Migrator()
	targetMigrator := s.targetDB.Migrator()

	// Check if table exists in both databases
	if !sourceMigrator.HasTable(entity) || !targetMigrator.HasTable(entity) {
		return false, nil
	}

	// Compare column definitions
	entityType := reflect.TypeOf(entity).Elem()
	for i := 0; i < entityType.NumField(); i++ {
		field := entityType.Field(i)
		columnName := s.getColumnName(field)
		if columnName == "" {
			continue // Skip non-database fields
		}

		// Compare column types
		sourceColType, _ := sourceMigrator.ColumnTypes(entity)
		targetColType, _ := targetMigrator.ColumnTypes(entity)
		if !s.compareColumnTypes(sourceColType, targetColType) {
			return false, nil
		}
	}
	return true, nil
}

func (s *DatabaseSynchronizer) getColumnName(field reflect.StructField) string {
	tagSettings := schema.ParseTagSetting(field.Tag.Get("gorm"), ";")
	if name, ok := tagSettings["COLUMN"]; ok {
		return name
	} else if name, ok := tagSettings["column"]; ok {
		return name
	}
	return field.Name
}

func (s *DatabaseSynchronizer) compareColumnTypes(source, target []gorm.ColumnType) bool {
	// Implement detailed column type comparison
	// This is a simplified version - expand based on your needs
	if len(source) != len(target) {
		return false
	}

	for i := range source {
		if source[i].Name() != target[i].Name() ||
			source[i].DatabaseTypeName() != target[i].DatabaseTypeName() {
			return false
		}
	}
	return true
}

func (s *DatabaseSynchronizer) getSourceRecords(entity interface{}) ([]interface{}, error) {
	var records []interface{}
	result := s.sourceDB.Model(entity).Find(&records)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to fetch records: %v", result.Error)
	}
	return records, nil
}
