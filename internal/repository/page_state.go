// Copyright (c) 2024
// Licensed under the MIT License. See LICENSE file in the root directory.

package repository

import (
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// How this machine shows a page: whether the folder is open and how far the
// reader got. Never synced.
type PageState struct {
	PageID    string `gorm:"primaryKey"`
	Expanded  bool
	ReadWord  int
	ReadTotal int
}

type ReadProgress struct {
	Word  int `json:"word"`
	Total int `json:"total"`
}

type PageStateRepository struct {
	BaseRepository
}

func NewPageStateRepository(unSyncedDB *gorm.DB) *PageStateRepository {
	return &PageStateRepository{BaseRepository: BaseRepository{db: unSyncedDB}}
}

func (r *PageStateRepository) Get(pageID string) PageState {
	state := PageState{PageID: pageID}
	r.db.First(&state, "page_id = ?", pageID)
	return state
}

func (r *PageStateRepository) All() map[string]PageState {
	var states []PageState
	r.db.Find(&states)

	byID := make(map[string]PageState, len(states))
	for _, state := range states {
		byID[state.PageID] = state
	}
	return byID
}

func (r *PageStateRepository) SetExpanded(pageID string, expanded bool) error {
	return r.upsert(&PageState{PageID: pageID, Expanded: expanded}, "expanded")
}

func (r *PageStateRepository) SaveReadProgress(pageID string, word, total int) error {
	return r.upsert(&PageState{PageID: pageID, ReadWord: word, ReadTotal: total}, "read_word", "read_total")
}

func (r *PageStateRepository) Delete(pageID string) error {
	return r.db.Delete(&PageState{}, "page_id = ?", pageID).Error
}

// Writes only the named columns so the other half of the row survives.
func (r *PageStateRepository) upsert(state *PageState, columns ...string) error {
	return r.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "page_id"}},
		DoUpdates: clause.AssignmentColumns(columns),
	}).Create(state).Error
}
