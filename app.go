package main

import (
	"context"
	"fmt"
	"log"
	"myscript/internal/notion"
	"myscript/internal/repository"
	"myscript/internal/transcribe/structs"
	witai "myscript/internal/transcribe/wait.ai"
	"myscript/internal/transcribe/whisper"
	"myscript/internal/transcribe/whisper/openai"
	"myscript/internal/utils"
	"myscript/internal/utils/microphone"

	"github.com/wailsapp/wails/v2/pkg/runtime"

	"github.com/jomei/notionapi"
	"gorm.io/gorm"
)

// App struct
type App struct {
	ctx            context.Context
	audioSequencer *microphone.AudioSequencer
	db             *gorm.DB
}

// NewApp creates a new App application struct
func NewApp(db *gorm.DB, audioSequencer *microphone.AudioSequencer) *App {
	return &App{db: db, audioSequencer: audioSequencer}
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

// --- WitAI ---

func (a *App) GetWitAILanguages() []structs.Language {
	return witai.GetSupportedLanguages()
}

// --- Languages ---

func (a *App) GetLanguages() []structs.Language {
	config := a.GetConfig()

	switch config.TranscriberSource {
	case "witai":
		return a.GetWitAILanguages()
	default:
		return a.GetWhisperLanguages()
	}
}

// --- Transcribe ---

func (a *App) WitTranscribe(buffer []byte, language string) (string, error) {
	apiKey := witai.GetAPIKey(language)
	if apiKey == nil {
		return "", fmt.Errorf("no API key found for language %s", language)
	}

	return witai.WitAITranscribeFromBuffer(buffer, apiKey.Key)
}

func (a *App) OpenAPITranscribe(buffer []byte, language string) (string, error) {
	config := a.GetConfig()

	if config.OpenAIApiKey == nil || *config.OpenAIApiKey == "" {
		return "", fmt.Errorf("no OpenAI API key found")
	}

	text, err := openai.TranscribeFromBuffer(buffer, language, *config.OpenAIApiKey)
	if err != nil {
		return "", err
	}

	return text, nil
}

func (a *App) LocalTranscribe(buffer []byte, language string) (string, error) {
	return "", fmt.Errorf("not implemented")
}

func (a *App) Transcribe(buffer []byte, language string) (string, error) {
	config := a.GetConfig()

	log.Printf("Transcribing with language: %s, source: %s", language, config.TranscriberSource)

	switch config.TranscriberSource {
	case "witai":
		return a.WitTranscribe(buffer, language)
	case "openai":
		return a.OpenAPITranscribe(buffer, language)
	case "local":
		return a.LocalTranscribe(buffer, language)
	}

	return "", fmt.Errorf("invalid transcriber source: %s", config.TranscriberSource)
}

// --- Microphone Recording ---

func (a *App) StartRecording(language string) error {
	log.Printf("Starting recording with language: %s", language)

	pq := utils.NewProcessQueue("transcriber-queue")

	a.audioSequencer.SetSequentializeCallback(func(buffer []byte) {
		bookId := pq.Book()

		waveBuffer, _ := a.audioSequencer.RawBytesToWAV(buffer)
		transcribed, err := a.Transcribe(waveBuffer, language)

		if err != nil {
			log.Printf("Transcription error: %s\n", err.Error())
			runtime.EventsEmit(a.ctx, "on-transcribe-error", err.Error())
			return
		}

		pq.Add(bookId, func() {
			runtime.EventsEmit(a.ctx, "on-transcribed-text", transcribed)
		})
	})

	a.audioSequencer.SetStopCallback(func(autoStopped bool) {
		runtime.EventsEmit(a.ctx, "on-recording-stopped", autoStopped)
	})

	return a.audioSequencer.Start()
}

func (a *App) StopRecording() {
	log.Println("Stopping recording")

	a.audioSequencer.Stop(false)
	// Should be called after Stop()
	a.audioSequencer.SetSequentializeCallback(nil)
	a.audioSequencer.SetStopCallback(nil)
}

func (a *App) IsRecording() bool {
	return a.audioSequencer.Recording()
}
