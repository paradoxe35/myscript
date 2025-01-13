package repository

import (
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type Page struct {
	gorm.Model
	Title       string         `json:"title"`
	HtmlContent string         `json:"html_content"`
	Blocks      datatypes.JSON `json:"blocks"`
	IsFolder    bool           `json:"is_folder"`
	Expanded    bool           `json:"expanded"`
	Order       int            `json:"order"`
	// Self-referential relationship
	ParentID *uint  `gorm:"index;constraint:OnDelete:SET NULL"` // Nullable, allows for root-level pages
	Children []Page `gorm:"foreignkey:ParentID"`
}

func GetPages(db *gorm.DB) []Page {
	var pages []Page

	db.Omit("html_content", "blocks").Find(&pages)

	return pages
}

func GetPage(db *gorm.DB, id uint) *Page {
	var page Page
	db.First(&page, id)

	return &page
}

func UpdatePageOrder(db *gorm.DB, id uint, ParentID *uint, order int) {
	db.Model(&Page{}).Where("id = ?", id).
		Update("order", order).
		Update("ParentID", ParentID)
}

func SavePage(db *gorm.DB, page *Page) *Page {
	db.Save(page)
	return page
}

func DeletePage(db *gorm.DB, id uint) {
	var page Page

	db.First(&page, id)
	db.Delete(&page)
}
