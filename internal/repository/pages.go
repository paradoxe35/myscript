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
	Expanded    bool           `json:"expanded"`
	Order       int            `json:"order"`
	// Self-referential relationship
	ParentID *string `gorm:"index"` // Nullable parent reference
	Children []Page  `gorm:"foreignKey:ParentID;constraint:OnDelete:SET NULL;"`
}

type MapUpdate = map[string]interface{}

func GetPages(db *gorm.DB) []Page {
	var pages []Page

	db.Omit("html_content", "blocks").Find(&pages)

	return pages
}

func GetPage(db *gorm.DB, ID string) *Page {
	var page Page
	db.First(&page, ID)

	return &page
}

func UpdatePageOrder(db *gorm.DB, ID string, ParentID *string, order int) {
	db.Model(&Page{}).
		Where("id = ?", ID).
		Updates(MapUpdate{
			"order":    order,
			"ParentID": ParentID,
		})
}

func SavePage(db *gorm.DB, page *Page) *Page {
	db.Save(page)
	return page
}

func DeletePage(db *gorm.DB, ID string) {
	var page Page

	db.First(&page, ID)
	db.Delete(&page)
}
