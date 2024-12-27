package local_whisper

import (
	"context"
	"fmt"
	"myscript/internal/filesystem"
	"myscript/internal/utils"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"syscall"
	"time"
)

const (
	srcUrl          = "https://huggingface.co/ggerganov/whisper.cpp/resolve/main?download=true" // The location of the models
	srcExt          = ".bin"                                                                    // Filename extension
	bufSize         = 1024 * 64                                                                 // Size of the buffer used for downloading the model
	outDir          = "./models"                                                                // Directory where the model will be downloaded
	downloadTimeout = 30 * time.Minute                                                          // Timeout for downloading the model
)

type LocalWhisperModel struct {
	Name        string
	EnglishOnly bool
}

type DownloadProgress struct {
	Name  string
	Size  int64
	Total int64
}

var modelsDownloadProgress = make(map[string]bool)

func ModelExists(model LocalWhisperModel) bool {
	modelsDir, err := getModelOutDir()
	if err != nil {
		return false
	}

	modelName := getModelFullName(model)

	modelPath := filepath.Join(modelsDir, modelName)

	info, err := os.Stat(modelPath)

	return err == nil && info != nil && !info.IsDir()
}

func DownloadModels(models []LocalWhisperModel, progress chan<- DownloadProgress) error {
	modelsDir, err := getModelOutDir()
	if err != nil {
		close(progress)
		return err
	}

	if areModelsDownloading(models) {
		close(progress)
		return nil
	}

	// Remove in progress models
	defer func() {
		for modelName := range modelsDownloadProgress {
			delete(modelsDownloadProgress, modelName)
		}
	}()

	putModelsDownloading(models)

	ctx := utils.ContextForSignal(os.Interrupt, os.Kill, syscall.SIGQUIT)

	for _, model := range models {
		modelName := getModelFullName(model)
		modelPath := filepath.Join(modelsDir, modelName)

		url, err := urlForModel(model)
		if err != nil {
			close(progress)
			return err
		}

		// Download the model
		err = downloadModel(ctx, url, modelPath, progress)
		if err != nil {
			os.Remove(modelPath)
			close(progress)
			return err
		}
	}

	close(progress)

	return nil
}

func AreSomeModelsDownloading() bool {
	return len(modelsDownloadProgress) > 0
}

func IsModelDownloading(model LocalWhisperModel) bool {
	modelName := getModelFullName(model)
	_, ok := modelsDownloadProgress[modelName]
	return ok
}

func putModelsDownloading(models []LocalWhisperModel) {
	for _, model := range models {
		modelName := getModelFullName(model)
		modelsDownloadProgress[modelName] = true
	}
}

func areModelsDownloading(models []LocalWhisperModel) bool {
	downloading := false

	for _, model := range models {
		modelName := getModelFullName(model)
		if _, ok := modelsDownloadProgress[modelName]; ok {
			downloading = true
			break
		}
	}

	return downloading
}

func downloadModel(ctx context.Context, url string, modelPath string, progress chan<- DownloadProgress) error {
	fmt.Printf("Downloading model: %s\n", url)

	// Create HTTP client
	client := http.Client{
		Timeout: downloadTimeout,
	}

	modelName := filepath.Base(modelPath)

	// Initiate the download
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("%s: %s", modelName, resp.Status)
	}

	// If output file exists and is the same size as the model, skip
	if info, err := os.Stat(modelPath); err == nil && info.Size() == resp.ContentLength {
		fmt.Println("Skipping", modelName, "as it already exists")
		return nil
	}

	// Create file
	w, err := os.Create(modelPath)
	if err != nil {
		return err
	}
	defer w.Close()

	// Report
	fmt.Println("Downloading", modelName, "to", modelPath)

	// Progressively download the model
	data := make([]byte, bufSize)
	count := int64(0)
	ticker := time.NewTicker(5 * time.Second)

	for {
		select {
		case <-ctx.Done():
			// Cancelled, return error
			return ctx.Err()
		case <-ticker.C:
			progress <- DownloadProgress{
				Name:  modelName,
				Size:  count,
				Total: resp.ContentLength,
			}
		default:
			// Read body
			n, err := resp.Body.Read(data)
			if err != nil {
				progress <- DownloadProgress{
					Name:  modelName,
					Size:  count,
					Total: resp.ContentLength,
				}
				return err
			} else if m, err := w.Write(data[:n]); err != nil {
				return err
			} else {
				count += int64(m)
			}
		}
	}
}

func getModelOutDir() (string, error) {
	dir := filepath.Join(filesystem.HOME_DIR, outDir)

	_, err := os.Stat(dir)
	if err != nil {
		if os.IsNotExist(err) {
			err = os.Mkdir(dir, 0755)
			if err != nil {
				return "", err
			}
		} else {
			return "", err
		}
	}

	return dir, nil
}

func getModelFullName(model LocalWhisperModel) string {
	modelName := model.Name
	if model.EnglishOnly {
		modelName += ".en"
	}

	if filepath.Ext(model.Name) != srcExt {
		modelName += srcExt
	}

	return modelName
}

// URLForModel returns the URL for the given model on huggingface.co
func urlForModel(model LocalWhisperModel) (string, error) {
	modelName := getModelFullName(model)

	url, err := url.Parse(srcUrl)
	if err != nil {
		return "", err
	} else {
		url.Path = url.Path + "/" + modelName
	}

	return url.String(), nil
}
