// Copyright (c) 2024
// Licensed under the MIT License. See LICENSE file in the root directory.

package repository

import (
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// UNSYNCED MODEL

type ProcessedChange struct {
	gorm.Model
	FileID string `gorm:"uniqueIndex"`
}

type ProcessedChangeRepository struct {
	BaseRepository
}

func NewProcessedChangeRepository(unSyncedDB *gorm.DB) *ProcessedChangeRepository {
	return &ProcessedChangeRepository{
		BaseRepository: BaseRepository{db: unSyncedDB},
	}
}

func (r *ProcessedChangeRepository) GetProcessedChange(fileID string) *ProcessedChange {
	var ProcessedChange ProcessedChange

	err := r.db.
		Or("file_id = ?", fileID).
		First(&ProcessedChange).Error

	if err != nil {
		return nil
	}

	return &ProcessedChange
}

func (r *ProcessedChangeRepository) ChangeProcessed(fileID string) bool {
	return r.GetProcessedChange(fileID) != nil
}

func (r *ProcessedChangeRepository) DeleteProcessedChange(fileID string) error {
	return r.db.
		Unscoped().
		Where("file_id = ?", fileID).
		Delete(&ProcessedChange{}).Error
}

func (r *ProcessedChangeRepository) DeleteOldProcessedChanges(timeOffset time.Time) error {
	return r.db.Unscoped().Where("created_at < ?", timeOffset).Delete(&ProcessedChange{}).Error
}

func (r *ProcessedChangeRepository) SaveProcessedChange(fileID string) error {
	// Save with a zero id inserts, so re-recording a file used to add a row
	// every cycle rather than leaving the one already there.
	return r.db.Clauses(clause.OnConflict{DoNothing: true}).
		Create(&ProcessedChange{FileID: fileID}).Error
}
