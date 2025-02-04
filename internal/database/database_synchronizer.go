package database

import (
	"encoding/json"
	"fmt"
	"myscript/internal/repository"
	"reflect"
	"sync"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/schema"
)

type AffectedTables map[string][]string

type DatabaseSynchronizer struct {
	sourceDB       *gorm.DB
	targetDB       *gorm.DB
	affectedTables AffectedTables
}

// Add these types to DatabaseSynchronizer
type SyncStrategy func(entity interface{}) error

type EntitySyncRule struct {
	OnConflictOmitColumns []string
	ConflictColumns       []string
	Strategy              SyncStrategy
}

func NewDatabaseSynchronizer(sourceDB *gorm.DB, targetDB *gorm.DB) *DatabaseSynchronizer {
	sourceDb := sourceDB
	targetDb := targetDB

	if sourceDB != nil {
		sourceDb = sourceDB.Session(&gorm.Session{SkipHooks: true})
	}

	if targetDB != nil {
		targetDb = targetDB.Session(&gorm.Session{SkipHooks: true})
	}

	return &DatabaseSynchronizer{
		sourceDB: sourceDb,
		targetDB: targetDb,
	}
}

func (s *DatabaseSynchronizer) SynchronizeAll() error {
	entities := []interface{}{
		&repository.Config{},
		&repository.Page{},
		&repository.Cache{},
	}

	for _, entity := range entities {
		if err := s.synchronizeSourceEntity(entity); err != nil {
			return fmt.Errorf("synchronization failed for %T: %v", entity, err)
		}
	}
	return nil
}

func (s *DatabaseSynchronizer) SynchronizeChangeLog(changeLog repository.ChangeLog) error {
	switch changeLog.Operation {
	case repository.OPERATION_SAVE: // Handle both CREATE and UPDATE
		var model interface{}

		switch changeLog.TableName {
		case s.GetEntityTableName(&repository.Config{}):
			model = &repository.Config{}
		case s.GetEntityTableName(&repository.Page{}):
			model = &repository.Page{}
		case s.GetEntityTableName(&repository.Cache{}):
			model = &repository.Cache{}
		default:
			return fmt.Errorf("unsupported table name: %s", changeLog.TableName)
		}

		// Unmarshal JSON data into the model
		if err := json.Unmarshal([]byte(changeLog.NewData), model); err != nil {
			return err
		}

		if err := s.synchronizeEntity(model, []interface{}{model}); err != nil {
			return err
		}

	case repository.OPERATION_DELETE:
		s.targetDB.
			Table(changeLog.TableName).
			Where("id = ?", changeLog.RowID).
			Delete(nil)
	}

	return nil
}

func (s *DatabaseSynchronizer) synchronizeEntity(entity interface{}, records []interface{}) error {
	return s.targetDB.Transaction(func(tx *gorm.DB) error {

		tableName := s.GetEntityTableName(entity)
		if tableName == "" {
			return fmt.Errorf("failed to get table name for %T", entity)
		}

		// Get sync rules based on entity type
		rules := s.getSyncRules(entity)

		// Apply custom sync strategy if exists
		if rules.Strategy != nil {
			for _, record := range records {
				if err := rules.Strategy(record); err != nil {
					return err
				}

				s.addAffectedTable(tableName, record)
			}
			return nil
		}

		// Default upsert behavior with conflict columns
		if len(rules.ConflictColumns) > 0 {
			for _, record := range records {
				err := tx.Table(tableName).
					Clauses(clause.OnConflict{
						Columns:   s.getConflictColumns(entity, rules.ConflictColumns),
						DoUpdates: clause.AssignmentColumns(s.getUpdateColumns(entity)),
					}).
					Omit(rules.OnConflictOmitColumns...).
					Create(record).Error

				if err != nil {
					return err
				}

				s.addAffectedTable(tableName, record)
			}
			return nil
		}

		// Fallback to basic save
		for _, record := range records {
			if err := tx.Table(tableName).Save(record).Error; err != nil {
				return err
			}

			s.addAffectedTable(tableName, record)
		}
		return nil
	})
}

func (s *DatabaseSynchronizer) GetAffectedTables() map[string][]string {
	return s.affectedTables
}

func (s *DatabaseSynchronizer) addAffectedTable(tableName string, record interface{}) {
	if s.affectedTables == nil {
		s.affectedTables = make(map[string][]string)
	}

	if _, ok := s.affectedTables[tableName]; !ok {
		s.affectedTables[tableName] = []string{}
	}

	recordId := repository.GetModelID(record)
	s.affectedTables[tableName] = append(s.affectedTables[tableName], recordId)
}

func (s *DatabaseSynchronizer) synchronizeSourceEntity(entity interface{}) error {
	// Schema compatibility check
	if compatible, err := s.areSchemasCompatible(entity); !compatible || err != nil {
		if err != nil {
			return fmt.Errorf("schema check error: %v", err)
		}
		return fmt.Errorf("schemas are incompatible for %T", entity)
	}

	records, err := s.getSourceRecords(entity)
	if err != nil {
		return err
	}

	return s.synchronizeEntity(entity, records)
}

func (s *DatabaseSynchronizer) getSyncRules(entity interface{}) EntitySyncRule {
	switch entity.(type) {
	case *repository.Config:
		return EntitySyncRule{
			Strategy: s.syncConfigStrategy,
		}
	case *repository.Cache:
		return EntitySyncRule{
			ConflictColumns:       []string{"key"},
			OnConflictOmitColumns: []string{"id"},
		}
	default:
		return EntitySyncRule{}
	}
}

func (s *DatabaseSynchronizer) GetEntityTableName(model interface{}) string {
	// Use the default naming strategy
	c, err := schema.Parse(model, &sync.Map{}, &schema.NamingStrategy{})
	if err != nil {
		return ""
	}
	return c.Table
}

func (s *DatabaseSynchronizer) areSchemasCompatible(entity interface{}) (bool, error) {
	sourceMigrator := s.sourceDB.Migrator()
	targetMigrator := s.targetDB.Migrator()

	// 1. Basic table existence check
	if !sourceMigrator.HasTable(entity) || !targetMigrator.HasTable(entity) {
		return false, nil
	}

	// 2. Get essential column information
	sourceCols, err := sourceMigrator.ColumnTypes(entity)
	if err != nil {
		return false, fmt.Errorf("failed to get source columns: %v", err)
	}

	targetCols, err := targetMigrator.ColumnTypes(entity)
	if err != nil {
		return false, fmt.Errorf("failed to get target columns: %v", err)
	}

	// 3. Check column compatibility
	for _, sCol := range sourceCols {
		var targetCol gorm.ColumnType = nil
		for _, tCol := range targetCols {
			if sCol.Name() == tCol.Name() {
				targetCol = tCol
				break
			}
		}

		if targetCol == nil || sCol.DatabaseTypeName() != targetCol.DatabaseTypeName() {
			return false, nil
		}
	}

	return true, nil
}

func (s *DatabaseSynchronizer) getSourceRecords(entity interface{}) ([]interface{}, error) {
	// Dereference pointers to get the underlying type
	entityValue := reflect.Indirect(reflect.ValueOf(entity))
	entityType := entityValue.Type()

	// Create a slice of pointers to the actual struct type
	sliceType := reflect.SliceOf(reflect.PtrTo(entityType))
	slicePtr := reflect.New(sliceType)

	// Execute query with proper typing
	result := s.sourceDB.Model(entity).Find(slicePtr.Interface())
	if result.Error != nil {
		return nil, fmt.Errorf("failed to fetch records: %v", result.Error)
	}

	// Convert to []interface{} with pointer elements
	slice := slicePtr.Elem()
	records := make([]interface{}, slice.Len())
	for i := 0; i < slice.Len(); i++ {
		// Get the pointer value and ensure it's addressable
		item := slice.Index(i)
		if item.Kind() == reflect.Ptr && item.IsNil() {
			continue // Skip nil pointers
		}
		records[i] = item.Interface()
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
func (s *DatabaseSynchronizer) getConflictColumns(entity interface{}, names []string) []clause.Column {
	columns := make([]clause.Column, len(names))

	// Create new statement to ensure schema is initialized
	stmt := &gorm.Statement{DB: s.targetDB}
	if err := stmt.Parse(entity); err != nil {
		return columns
	}

	for i, name := range names {
		if field := stmt.Schema.LookUpField(name); field != nil {
			columns[i] = clause.Column{Name: field.DBName}
		}
	}
	return columns
}

// Helper to get updatable columns (exclude conflict columns and primary keys)
func (s *DatabaseSynchronizer) getUpdateColumns(entity interface{}) []string {
	// Create a temporary statement to parse the entity
	stmt := &gorm.Statement{DB: s.targetDB}
	if err := stmt.Parse(entity); err != nil {
		fmt.Printf("Error parsing schema for update columns: %v", err)
		return nil
	}

	var columns []string
	conflictColumns := s.getSyncRules(entity).ConflictColumns

	for _, field := range stmt.Schema.Fields {
		// Skip primary keys and conflict columns
		if field.PrimaryKey || contains(conflictColumns, field.Name) {
			continue
		}

		// Skip created_at timestamp
		if field.Name == "CreatedAt" {
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
