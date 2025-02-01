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
	return repository.NewPageRepository().
		GetPages(a.syncedDb)
}

func (a *App) GetLocalPage(ID string) *repository.Page {
	return repository.NewPageRepository().
		GetPage(a.syncedDb, ID)
}

func (a *App) SaveLocalPage(page *repository.Page) *repository.Page {
	return repository.NewPageRepository().
		SavePage(a.syncedDb, page)
}

func (a *App) DeleteLocalPage(ID string) {
	repository.NewPageRepository().
		DeletePage(a.syncedDb, ID)
}

func (a *App) UpdateLocalPageOrder(ID string, ParentID *string, order int) {
	repository.NewPageRepository().
		UpdatePageOrder(a.syncedDb, ID, ParentID, order)
}
