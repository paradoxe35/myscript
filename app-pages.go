package main

import (
	"myscript/internal/notion"
	"myscript/internal/repository"

	"github.com/jomei/notionapi"
)

// --- Notion Pages ---

func (a *App) getNotionClient() *notion.NotionClient {
	config := a.GetConfig()
	if config == nil || config.NotionApiKey == nil {
		return nil
	}

	return notion.NewClient(*config.NotionApiKey)
}

func (a *App) GetNotionPages() ([]notionapi.Object, error) {
	client := a.getNotionClient()
	if client == nil {
		return []notionapi.Object{}, nil
	}

	return client.GetPages()
}

func (a *App) GetNotionPageBlocks(pageID string) ([]*notion.NotionBlock, error) {
	client := a.getNotionClient()

	return client.GetPageBlocks(pageID)
}

// --- Local Pages ---

func (a *App) GetLocalPages() []repository.Page {
	return repository.GetPages(a.db)
}

func (a *App) GetLocalPage(id uint) *repository.Page {
	return repository.GetPage(a.db, id)
}

func (a *App) SaveLocalPage(page *repository.Page) *repository.Page {
	return repository.SavePage(a.db, page)
}

func (a *App) DeleteLocalPage(id uint) {
	repository.DeletePage(a.db, id)
}

func (a *App) UpdateLocalPageOrder(id uint, ParentID *uint, order int) {
	repository.UpdatePageOrder(a.db, id, ParentID, order)
}
