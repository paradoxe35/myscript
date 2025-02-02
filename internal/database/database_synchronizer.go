package database

import (
	"fmt"
	"myscript/internal/repository"
	"reflect"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/schema"
)

type DatabaseSynchronizer struct {
	sourceDB *gorm.DB
	targetDB *gorm.DB
}

// Add these types to DatabaseSynchronizer
type SyncStrategy func(entity interface{}) error

type EntitySyncRule struct {
	ConflictColumns []string
	Strategy        SyncStrategy
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

func (s *DatabaseSynchronizer) getSyncRules(entity interface{}) EntitySyncRule {
	switch entity.(type) {
	case *repository.Config:
		return EntitySyncRule{
			Strategy: s.syncConfigStrategy,
		}
	case *repository.Page:
		return EntitySyncRule{}
	case *repository.Cache:
		return EntitySyncRule{
			ConflictColumns: []string{"key"},
		}
	default:
		return EntitySyncRule{}
	}
}

func (s *DatabaseSynchronizer) SynchronizeEntity(entity interface{}) error {
	// Get sync rules based on entity type
	rules := s.getSyncRules(entity)

	// Schema compatibility check
	if compatible, err := s.areSchemasCompatible(entity); !compatible || err != nil {
		if err != nil {
			return fmt.Errorf("schema check error: %v", err)
		}
		return fmt.Errorf("schemas are incompatible for %T", entity)
	}

	return s.targetDB.Transaction(func(tx *gorm.DB) error {
		records, err := s.getSourceRecords(entity)
		if err != nil {
			return err
		}

		// Apply custom sync strategy if exists
		if rules.Strategy != nil {
			for _, record := range records {
				if err := rules.Strategy(record); err != nil {
					return err
				}
			}
			return nil
		}

		// Default upsert behavior with conflict columns
		if len(rules.ConflictColumns) > 0 {
			return tx.Clauses(clause.OnConflict{
				Columns:   s.getConflictColumns(entity, rules.ConflictColumns),
				DoUpdates: clause.AssignmentColumns(s.getUpdateColumns(entity)),
			}).Create(records).Error
		}

		// Fallback to basic save
		return tx.Save(records).Error
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

// Custom strategy for Config entity (single row management)
func (s *DatabaseSynchronizer) syncConfigStrategy(entity interface{}) error {
	config, ok := entity.(*repository.Config)
	if !ok {
		return fmt.Errorf("invalid type for config strategy")
	}

	// Get existing config if exists
	var existing repository.Config
	if err := s.targetDB.First(&existing).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			// Create new config if none exists
			return s.targetDB.Create(config).Error
		}
		return err
	}

	// Update existing config
	return s.targetDB.Model(&existing).Updates(config).Error
}

// Helper to get conflict columns as clauses
func (s *DatabaseSynchronizer) getConflictColumns(_ interface{}, names []string) []clause.Column {
	columns := make([]clause.Column, len(names))
	st := s.targetDB.Statement.Schema
	for i, name := range names {
		if field := st.LookUpField(name); field != nil {
			columns[i] = clause.Column{Name: field.DBName}
		}
	}
	return columns
}

// Helper to get updatable columns (exclude conflict columns and primary keys)
func (s *DatabaseSynchronizer) getUpdateColumns(entity interface{}) []string {
	st := s.targetDB.Statement.Schema
	var columns []string

	for _, field := range st.Fields {
		// Skip primary keys and conflict columns
		if field.PrimaryKey || contains(s.getSyncRules(entity).ConflictColumns, field.Name) {
			continue
		}
		columns = append(columns, field.DBName)
	}
	return columns
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
