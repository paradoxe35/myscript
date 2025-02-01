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

// Functions

func (r *SyncStateRepository) GetSyncState() *SyncState {
	var syncState SyncState

	r.db.First(&syncState)

	return &syncState
}

func (r *SyncStateRepository) SaveSyncState(syncTimeOffset time.Time) *SyncState {
	syncState := r.GetSyncState()
	syncState.SyncTimeOffset = syncTimeOffset
	r.db.Save(syncState)

	return syncState
}
