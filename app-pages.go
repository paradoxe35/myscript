// Copyright (c) 2024
// Licensed under the MIT License. See LICENSE file in the root directory.

package main

import (
	"errors"
	"myscript/internal/notion"
	"myscript/internal/repository"

	"github.com/jomei/notionapi"
)

func (a *App) localCache() *repository.LocalCacheRepository {
	return repository.NewLocalCacheRepository(a.unSyncedDB)
}

func (a *App) pages() *repository.PageRepository {
	return repository.NewPageRepository(a.mainDB, a.unSyncedDB)
}

func (a *App) pageStates() *repository.PageStateRepository {
	return repository.NewPageStateRepository(a.unSyncedDB)
}

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

func notionBlocksKey(pageID string) string {
	return "notion_page:" + pageID
}

// Fetches the tree and keeps a copy for the next open.
func (a *App) GetNotionPageBlocks(pageID string) ([]*notion.NotionBlock, error) {
	client := a.getNotionClient()
	if client == nil {
		return nil, errors.New("no Notion API key configured")
	}

	blocks, err := client.GetPageBlocks(pageID)
	if err != nil {
		return nil, err
	}

	a.localCache().Set(notionBlocksKey(pageID), blocks)
	return blocks, nil
}

// The copy from the last fetch, or nil when the page was never opened.
func (a *App) GetCachedNotionPageBlocks(pageID string) []*notion.NotionBlock {
	var blocks []*notion.NotionBlock
	if !a.localCache().Get(notionBlocksKey(pageID), &blocks) {
		return nil
	}
	return blocks
}

func (a *App) GetLocalPages() []repository.Page {
	return a.pages().GetPages()
}

func (a *App) GetLocalPage(ID string) *repository.Page {
	return a.pages().GetPage(ID)
}

// The editor's full save; renames and folder toggles have their own calls.
func (a *App) SaveLocalPage(page *repository.Page) *repository.Page {
	return a.pages().SavePage(page)
}

func (a *App) UpdateLocalPageTitle(ID, title string) *repository.Page {
	return a.pages().UpdatePageTitle(ID, title)
}

func (a *App) DeleteLocalPage(ID string) {
	a.pages().DeletePage(ID)
}

func (a *App) UpdateLocalPageOrder(ID string, ParentID *string, order int) {
	a.pages().UpdatePageOrder(ID, ParentID, order)
}

func (a *App) SetPageExpanded(ID string, expanded bool) error {
	return a.pages().SetExpanded(ID, expanded)
}

func (a *App) GetPageReadProgress(ID string) repository.ReadProgress {
	state := a.pageStates().Get(ID)
	return repository.ReadProgress{Word: state.ReadWord, Total: state.ReadTotal}
}

func (a *App) SavePageReadProgress(ID string, word, total int) error {
	return a.pageStates().SaveReadProgress(ID, word, total)
}
