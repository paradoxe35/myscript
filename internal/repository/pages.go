package repository

import (
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type Page struct {
	gorm.Model
	Title  string         `json:"title"`
	Blocks datatypes.JSON `json:"blocks"`
}

func GetPages(db *gorm.DB) []Page {
	var pages []Page
	db.Select([]string{"id", "title"}).Find(&pages)

	return pages
}

func GetPage(db *gorm.DB, id uint) *Page {
	var page Page
	db.First(&page, id)

	return &page
}

func SavePage(db *gorm.DB, page *Page) {
	db.Save(page)
}
