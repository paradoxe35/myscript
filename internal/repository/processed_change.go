// Copyright (c) 2024
// Licensed under the MIT License. See LICENSE file in the root directory.

package repository

import (
	"time"

	"gorm.io/gorm"
)

// UNSYNCED MODEL

type ProcessedChange struct {
	gorm.Model
	FileID string `gorm:"index"`
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
	item := &ProcessedChange{
		FileID: fileID,
	}

	return r.db.Save(item).Error
}
