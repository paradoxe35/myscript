package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"myscript/internal/notion"
	"myscript/internal/repository"
	"myscript/internal/transcribe/structs"
	witai "myscript/internal/transcribe/wait.ai"
	"myscript/internal/transcribe/whisper"
	"myscript/internal/transcribe/whisper/openai"
	"myscript/internal/utils"

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
		return whisper.LOCAL_WHISPER_MODELS[0].Name
	}

	return whisper.SuggestWhisperModel(availableRAM)
}

func (a *App) GetWhisperLanguages() []structs.Language {
	return whisper.GetWhisperLanguages()
}

func (a *App) GetWhisperModels() []whisper.WhisperModel {
	return whisper.GetWhisperModels()
}

// --- Transcribe ---

func (a *App) WitTranscribe(base64Data string, language string) (string, error) {
	apiKey := witai.GetAPIKey(language)
	if apiKey == nil {
		return "", fmt.Errorf("no API key found for language %s", language)
	}

	data, err := base64.StdEncoding.DecodeString(base64Data)
	if err != nil {
		fmt.Println("Error decoding base64 data:", err)
		return "", err
	}

	return witai.WitAITranscribeFromBuffer(data, apiKey.Key)
}

func (a *App) OpenAPITranscribe(base64Data string, language string) (string, error) {
	config := a.GetConfig()

	if config.OpenAIApiKey == nil || *config.OpenAIApiKey == "" {
		return "", fmt.Errorf("no OpenAI API key found")
	}

	data, err := base64.StdEncoding.DecodeString(base64Data)
	if err != nil {
		fmt.Println("Error decoding base64 data:", err)
		return "", err
	}

	text, err := openai.TranscribeFromBuffer(data, language, *config.OpenAIApiKey)
	if err != nil {
		return "", err
	}

	return text, nil
}
