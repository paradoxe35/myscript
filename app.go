package main

import (
	"context"
	"myscript/internal/notion"
	"myscript/internal/repository"
	"myscript/internal/utils"
	"myscript/internal/whisper"

	"github.com/jomei/notionapi"
	"gorm.io/gorm"
)

// App struct
type App struct {
	ctx context.Context
	db  *gorm.DB
}

// NewApp creates a new App application struct
func NewApp(db *gorm.DB) *App {
	return &App{db: db}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// --- Config ---

func (a *App) GetConfig() *repository.Config {
	return repository.GetConfig(a.db)
}

func (a *App) SaveConfig(config *repository.Config) *repository.Config {
	repository.SaveConfig(a.db, config)
	return a.GetConfig()
}

// --- Notion Pages ---

func (a *App) getNotionClient() *notion.NotionClient {
	config := a.GetConfig()
	if config == nil || config.NotionApiKey == nil {
		return nil
	}

	return notion.NewClient(*config.NotionApiKey)
}

func (a *App) GetNotionPages() []notionapi.Object {
	client := a.getNotionClient()
	if client == nil {
		return []notionapi.Object{}
	}

	return client.GetPages()
}

func (a *App) GetNotionPageBlocks(pageID string) []notion.NotionBlock {
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

// --- Cache ---

func (a *App) GetCache(key string) *repository.CacheValue {
	return repository.GetCache(a.db, key)
}

func (a *App) SaveCache(key string, value interface{}) *repository.Cache {
	return repository.SaveCache(a.db, key, value)
}

func (a *App) DeleteCache(key string) {
	repository.DeleteCache(a.db, key)
}

// --- Whisper ---

func (a *App) GetBestWhisperModel() string {
	availableRAM, err := utils.GetAvailableRAM()
	if err != nil {
		return "tiny"
	}

	return whisper.SuggestWhisperModel(availableRAM)
}
