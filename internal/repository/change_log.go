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

func LogChange(tx *gorm.DB, model interface{}, operation string) error {
	newData, _ := json.Marshal(model)

	change := ChangeLog{
		TableName: tx.Statement.Table,
		RowID:     getModelID(model),
		Operation: operation,
		NewData:   datatypes.JSON(newData),
		Synced:    false,
	}

	return tx.Create(&change).Error
}

func GetUnSyncedChanges(db *gorm.DB) []ChangeLog {
	var changes []ChangeLog

	db.Where("synced = ?", false).Find(&changes)

	return changes
}
