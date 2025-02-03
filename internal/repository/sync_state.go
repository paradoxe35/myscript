package repository

import (
	"time"

	"gorm.io/gorm"
)

// UNSYNCED MODEL

type SyncState struct {
	gorm.Model
	SyncTimeOffset time.Time
}

type SyncStateRepository struct {
	BaseRepository
}

func NewSyncStateRepository(unSyncedDB *gorm.DB) *SyncStateRepository {
	return &SyncStateRepository{
		BaseRepository: BaseRepository{db: unSyncedDB},
	}
}

var DEFAULT_TIME = time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)

func (r *SyncStateRepository) GetSyncState() *SyncState {
	var syncState SyncState

	if r.db.First(&syncState).Error != nil {
		syncState.SyncTimeOffset = DEFAULT_TIME
		r.db.Save(&syncState)
	}

	return &syncState
}

func (r *SyncStateRepository) SaveSyncState(syncTimeOffset time.Time) *SyncState {
	syncState := r.GetSyncState()

	if syncTimeOffset.IsZero() || syncTimeOffset.Before(DEFAULT_TIME) {
		syncState.SyncTimeOffset = DEFAULT_TIME
	} else {
		syncState.SyncTimeOffset = syncTimeOffset.Truncate(time.Second)
	}

	r.db.Save(syncState)

	return syncState
}
