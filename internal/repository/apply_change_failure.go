// Copyright (c) 2024
// Licensed under the MIT License. See LICENSE file in the root directory.

package repository

import (
	"time"

	"gorm.io/gorm"
)

// UNSYNCED MODEL

type RemoteApplyFailure struct {
	gorm.Model
	FileID string
	Count  int
}

type RemoteApplyFailureRepository struct {
	BaseRepository
}

func NewRemoteApplyFailureRepository(unSyncedDB *gorm.DB) *RemoteApplyFailureRepository {
	return &RemoteApplyFailureRepository{
		BaseRepository: BaseRepository{db: unSyncedDB},
	}
}

func (r *RemoteApplyFailureRepository) GetRemoteApplyFailure(fileID string) *RemoteApplyFailure {
	var remoteApplyFailure RemoteApplyFailure

	r.db.Where("file_id = ?", fileID).First(&remoteApplyFailure)

	return &remoteApplyFailure
}

func (r *RemoteApplyFailureRepository) DeleteOldRemoteApplyFailures(timeOffset time.Time) error {
	return r.db.Unscoped().Where("created_at < ?", timeOffset).Delete(&RemoteApplyFailure{}).Error
}

func (r *RemoteApplyFailureRepository) SaveRemoteApplyFailure(fileID string, count int) *RemoteApplyFailure {
	item := r.GetRemoteApplyFailure(fileID)

	item.FileID = fileID
	item.Count = count
	r.db.Save(item)

	return item
}
