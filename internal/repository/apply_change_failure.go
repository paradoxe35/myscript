package repository

import (
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

func NewRemoteApplyFailureRepository(unsyncedDb *gorm.DB) *RemoteApplyFailureRepository {
	return &RemoteApplyFailureRepository{
		BaseRepository: BaseRepository{db: unsyncedDb},
	}
}

func (r *RemoteApplyFailureRepository) GetRemoteApplyFailure(fileID string) *RemoteApplyFailure {
	var remoteApplyFailure RemoteApplyFailure

	r.db.Where("file_id = ?", fileID).First(&remoteApplyFailure)

	return &remoteApplyFailure
}

func (r *RemoteApplyFailureRepository) SaveRemoteApplyFailure(fileID string, count int) *RemoteApplyFailure {
	item := r.GetRemoteApplyFailure(fileID)

	item.FileID = fileID
	item.Count = count
	r.db.Save(item)

	return item
}
