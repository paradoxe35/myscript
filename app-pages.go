// Copyright (c) 2024
// Licensed under the MIT License. See LICENSE file in the root directory.

package main

import (
	"errors"
	"myscript/internal/notion"
	"myscript/internal/repository"

	"github.com/jomei/notionapi"
)

func (a *App) getNotionClient() *notion.NotionClient {
	apiKey := a.notionAPIKey()
	if apiKey == "" {
		return nil
	}

	return notion.NewClient(apiKey)
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
	if client == nil {
		return nil, errors.New("no Notion API key configured")
	}

	return client.GetPageBlocks(pageID)
}

func (a *App) GetLocalPages() []repository.Page {
	return repository.NewPageRepository(a.mainDB).
		GetPages()
}

func (a *App) GetLocalPage(ID string) *repository.Page {
	return repository.NewPageRepository(a.mainDB).
		GetPage(ID)
}

func (a *App) SaveLocalPage(page *repository.Page) *repository.Page {
	return repository.NewPageRepository(a.mainDB).
		SavePage(page)
}

func (a *App) DeleteLocalPage(ID string) {
	repository.NewPageRepository(a.mainDB).
		DeletePage(ID)
}

func (a *App) UpdateLocalPageOrder(ID string, ParentID *string, order int) {
	repository.NewPageRepository(a.mainDB).
		UpdatePageOrder(ID, ParentID, order)
}
