package repository

import (
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// !SYNCED MODEL

type Page struct {
	BaseUUIDModel

	Title       string         `json:"title"`
	HtmlContent string         `json:"html_content"`
	Blocks      datatypes.JSON `json:"blocks"`
	IsFolder    bool           `json:"is_folder"`
	Expanded    bool           `json:"expanded"`
	Order       int            `json:"order"`
	// Self-referential relationship
	ParentID *string `gorm:"index"` // Nullable parent reference
	Children []Page  `gorm:"foreignKey:ParentID;constraint:OnDelete:SET NULL;"`
}

// Hooks
func (n *Page) AfterCreate(tx *gorm.DB) error {
	return logChange(tx, n, OPERATION_CREATE)
}

func (n *Page) AfterUpdate(tx *gorm.DB) error {
	return logChange(tx, n, OPERATION_UPDATE)
}

func (n *Page) AfterDelete(tx *gorm.DB) error {
	return logChange(tx, n, OPERATION_DELETE)
}

type PageRepository struct {
	BaseRepository
}

func NewPageRepository(db *gorm.DB) *PageRepository {
	return &PageRepository{
		BaseRepository: BaseRepository{db: db},
	}
}

func (r *PageRepository) GetPages() []Page {
	var pages []Page

	r.db.Omit("html_content", "blocks").Find(&pages)

	return pages
}

func (r *PageRepository) GetPage(ID string) *Page {
	var page Page

	r.db.First(&page, "id = ?", ID)

	return &page
}

func (r *PageRepository) UpdatePageOrder(ID string, ParentID *string, order int) {
	r.db.Model(&Page{}).
		Where("id = ?", ID).
		Updates(MapUpdate{
			"order":    order,
			"ParentID": ParentID,
		})
}

func (r *PageRepository) SavePage(page *Page) *Page {
	r.db.Save(page)
	return page
}

func (r *PageRepository) DeletePage(ID string) {
	var page Page

	r.db.First(&page, "id = ?", ID)
	r.db.Delete(&page)
}
