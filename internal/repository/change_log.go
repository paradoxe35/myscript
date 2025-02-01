package repository

import (
	"encoding/json"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// UNSYNCED MODEL

type ChangeLog struct {
	gorm.Model
	TableName string
	RowID     string
	Operation string
	NewData   datatypes.JSON
	Synced    bool `gorm:"default:false"`
}

type ChangeLogRepository struct {
	BaseRepository
}

func NewChangeLogRepository(db *gorm.DB) *ChangeLogRepository {
	return &ChangeLogRepository{
		BaseRepository: BaseRepository{db: db},
	}
}

func logChange(tx *gorm.DB, model interface{}, operation string) error {
	newData, _ := json.Marshal(model)

	change := ChangeLog{
		TableName: tx.Statement.Table,
		RowID:     getModelID(model),
		Operation: operation,
		NewData:   datatypes.JSON(newData),
		Synced:    false,
	}

	// Check if the change already exists
	var existing ChangeLog
	if tx.Where("table_name = ? AND row_id = ? and operation = ?", change.TableName, change.RowID, operation).
		First(&existing).Error == nil {
		change.ID = existing.ID
		return tx.Save(&change).Error
	}

	return tx.Create(&change).Error
}

func (r *ChangeLogRepository) GetUnSyncedChanges() []ChangeLog {
	var changes []ChangeLog

	r.db.Where("synced = ?", false).Find(&changes)

	return changes
}
