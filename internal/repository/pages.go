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
	return logChange(tx, n, "CREATE")
}

func (n *Page) AfterUpdate(tx *gorm.DB) error {
	return logChange(tx, n, "UPDATE")
}

func (n *Page) AfterDelete(tx *gorm.DB) error {
	return logChange(tx, n, "DELETE")
}

type PageRepository struct {
	BaseRepository
}

func NewPageRepository() *PageRepository {
	return &PageRepository{}
}

func (*PageRepository) GetPages(db *gorm.DB) []Page {
	var pages []Page

	db.Omit("html_content", "blocks").Find(&pages)

	return pages
}

func (*PageRepository) GetPage(db *gorm.DB, ID string) *Page {
	var page Page

	db.First(&page, "id = ?", ID)

	return &page
}

func (*PageRepository) UpdatePageOrder(db *gorm.DB, ID string, ParentID *string, order int) {
	db.Model(&Page{}).
		Where("id = ?", ID).
		Updates(MapUpdate{
			"order":    order,
			"ParentID": ParentID,
		})
}

func (*PageRepository) SavePage(db *gorm.DB, page *Page) *Page {
	db.Save(page)
	return page
}

func (*PageRepository) DeletePage(db *gorm.DB, ID string) {
	var page Page

	db.First(&page, "id = ?", ID)
	db.Delete(&page)
}
