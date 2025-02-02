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

func NewSyncStateRepository(unsyncedDb *gorm.DB) *SyncStateRepository {
	return &SyncStateRepository{
		BaseRepository: BaseRepository{db: unsyncedDb},
	}
}

func (r *SyncStateRepository) GetSyncState() *SyncState {
	var syncState SyncState

	if r.db.First(&syncState).Error != nil {
		syncState.SyncTimeOffset = time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)
		r.db.Save(&syncState)
	}

	return &syncState
}

func (r *SyncStateRepository) SaveSyncState(syncTimeOffset time.Time) *SyncState {
	syncState := r.GetSyncState()
	syncState.SyncTimeOffset = syncTimeOffset
	r.db.Save(syncState)

	return syncState
}
