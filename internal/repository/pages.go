// Copyright (c) 2024
// Licensed under the MIT License. See LICENSE file in the root directory.

package repository

import (
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type Page struct {
	BaseUUIDModel

	Title       string         `json:"title"`
	HtmlContent string         `json:"html_content"`
	Blocks      datatypes.JSON `json:"blocks"`
	IsFolder    bool           `json:"is_folder"`
	Order       int            `json:"order"`
	ParentID    *string        `gorm:"index"`
	Children    []Page         `gorm:"foreignKey:ParentID;constraint:OnDelete:SET NULL;"`

	// Per machine, read from PageState; never written with the page.
	Expanded bool `json:"expanded" gorm:"-"`
}

func (n *Page) AfterCreate(tx *gorm.DB) error {
	return logChange(tx, n, OPERATION_SAVE)
}

func (n *Page) AfterUpdate(tx *gorm.DB) error {
	return logChange(tx, n, OPERATION_SAVE)
}

func (n *Page) AfterDelete(tx *gorm.DB) error {
	return logChange(tx, n, OPERATION_DELETE)
}

type PageRepository struct {
	BaseRepository
	states *PageStateRepository
}

func NewPageRepository(mainDB, unSyncedDB *gorm.DB) *PageRepository {
	return &PageRepository{
		BaseRepository: BaseRepository{db: mainDB},
		states:         NewPageStateRepository(unSyncedDB),
	}
}

func (r *PageRepository) GetPages() []Page {
	var pages []Page
	r.db.Omit("html_content", "blocks").Find(&pages)

	states := r.states.All()
	for i := range pages {
		pages[i].Expanded = states[pages[i].ID].Expanded
	}
	return pages
}

func (r *PageRepository) GetPage(ID string) *Page {
	var page Page
	r.db.First(&page, "id = ?", ID)

	page.Expanded = r.states.Get(ID).Expanded
	return &page
}

func (r *PageRepository) SavePage(page *Page) *Page {
	r.db.Save(page)
	return r.GetPage(page.ID)
}

// Partial updates go through the loaded row so the change log still carries
// the whole page.
func (r *PageRepository) UpdatePageTitle(ID, title string) *Page {
	r.db.Model(r.GetPage(ID)).
		Where("id = ?", ID).
		Updates(MapUpdate{"title": title})

	return r.GetPage(ID)
}

func (r *PageRepository) UpdatePageOrder(ID string, ParentID *string, order int) {
	r.db.Model(r.GetPage(ID)).
		Where("id = ?", ID).
		Updates(MapUpdate{
			"order":    order,
			"ParentID": ParentID,
		})
}

func (r *PageRepository) SetExpanded(ID string, expanded bool) error {
	return r.states.SetExpanded(ID, expanded)
}

func (r *PageRepository) DeletePage(ID string) {
	var page Page

	r.db.First(&page, "id = ?", ID)
	r.db.Delete(&page)
	r.states.Delete(ID)
}
