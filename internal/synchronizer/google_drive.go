package synchronizer

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"myscript/internal/repository"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"

	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
)

const PARENT_FOLDER = "appDataFolder"

type GoogleDriveService struct {
	ctx     context.Context
	service *drive.Service
}

func NewGoogleDriveService(client *http.Client) (*GoogleDriveService, error) {
	ctx := context.Background()

	service, err := drive.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil, err
	}

	return &GoogleDriveService{
		ctx:     ctx,
		service: service,
	}, nil
}

func (s *GoogleDriveService) createFileFromGoogleDrive(file *drive.File) *File {
	if file == nil {
		return nil
	}

	return &File{
		ID:          file.Id,
		Name:        file.Name,
		IsSnapshot:  strings.HasPrefix(file.Name, DB_SNAPSHOT_PREFIX),
		CreatedTime: s.parseTime(file.CreatedTime),
	}
}

func (s *GoogleDriveService) formatTime(t time.Time) string {
	return t.Format(time.RFC3339)
}

func (s *GoogleDriveService) parseTime(t string) time.Time {
	parsedTime, _ := time.Parse(time.RFC3339, t)
	return parsedTime
}

func (s *GoogleDriveService) files() *drive.FilesListCall {
	return s.service.Files.List().
		Spaces(PARENT_FOLDER)
}

func (s *GoogleDriveService) SortFilesAsc(files []*drive.File) []*drive.File {
	sort.Slice(files, func(i, j int) bool {
		timeI, _ := time.Parse(time.RFC3339, files[i].CreatedTime)
		timeJ, _ := time.Parse(time.RFC3339, files[j].CreatedTime)
		return timeI.Before(timeJ)
	})

	return files
}

func (s *GoogleDriveService) SortFilesDesc(files []*drive.File) []*drive.File {
	sort.Slice(files, func(i, j int) bool {
		timeI, _ := time.Parse(time.RFC3339, files[i].CreatedTime)
		timeJ, _ := time.Parse(time.RFC3339, files[j].CreatedTime)
		return timeI.After(timeJ)
	})

	return files
}

func (s *GoogleDriveService) geFileContent(fileId string) ([]byte, error) {
	resp, err := s.service.Files.Get(fileId).Download()
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if content, err := io.ReadAll(resp.Body); err != nil {
		return nil, err
	} else {
		return content, nil
	}
}

// func (s *GoogleDriveService) geFileContentStream(fileId string) (io.ReadCloser, error) {
// 	resp, err := s.service.Files.Get(fileId).Download()
// 	if err != nil {
// 		return nil, err
// 	}
// 	return resp.Body, nil
// }

func (s *GoogleDriveService) createFileWithJson(name string, content interface{}) (*drive.File, error) {
	file := &drive.File{
		Name:    name,
		Parents: []string{PARENT_FOLDER},
	}

	data, err := json.Marshal(content)
	if err != nil {
		slog.Error("GoogleDriveService[createFileWithJson] Marshal content to JSON", "error", err)
		return nil, err
	}

	if newFile, err := s.service.Files.Create(file).Media(bytes.NewReader(data)).Do(); err != nil {
		slog.Error("GoogleDriveService[createFileWithJson] Create file", "error", err)
		return nil, err
	} else {
		return newFile, nil
	}
}

func (s *GoogleDriveService) createRawFile(name string, content io.ReadSeeker) (*drive.File, error) {
	file := &drive.File{
		Name:    name,
		Parents: []string{PARENT_FOLDER},
	}

	if newFile, err := s.service.Files.Create(file).Media(content).Do(); err != nil {
		slog.Error("GoogleDriveService[createRawFile] Create file", "error", err)
		return nil, err
	} else {
		return newFile, nil
	}
}

func (s *GoogleDriveService) GetLatestDBSnapshot() (*File, error) {
	resp, err := s.files().
		PageSize(1).
		OrderBy("createdTime desc").
		Q(fmt.Sprintf("name contains '%s'", DB_SNAPSHOT_PREFIX)).
		Do()

	if err != nil {
		slog.Error("GoogleDriveService[GetLatestDBSnapshot] Get list DB_SNAPSHOT_PREFIX", "error", err)
		return nil, err
	}

	if len(resp.Files) == 0 {
		slog.Error("GoogleDriveService[GetLatestDBSnapshot] No DB snapshot found")
		return nil, fmt.Errorf("no DB snapshot found")
	}

	return s.createFileFromGoogleDrive(resp.Files[0]), nil
}

func (s *GoogleDriveService) SaveDBSnapshot(content io.ReadSeeker) (*File, error) {
	fileName := snapshotFileName(time.Now())
	file, err := s.createRawFile(fileName, content)

	if err != nil {
		slog.Error("GoogleDriveService[SaveDBSnapshot] Create file", "fileName", fileName, "error", err)
		return nil, err
	}

	return s.createFileFromGoogleDrive(file), nil
}

func (s *GoogleDriveService) GetChangeFilesAfterTimeOffset(timeOffset time.Time) ([]*File, error) {
	resp, err := s.files().
		OrderBy("createdTime asc").
		Q(fmt.Sprintf("createdTime > '%s'", s.formatTime(timeOffset))).
		Do()

	if err != nil {
		slog.Error("GoogleDriveService[GetChangeFilesAfterTimeOffset] Get list CHANGES_FILE_PREFIX", "error", err)
		return nil, err
	}

	files := make([]*File, len(resp.Files))
	for i, file := range resp.Files {
		files[i] = s.createFileFromGoogleDrive(file)
	}
	return files, nil
}

func (s *GoogleDriveService) UploadChangeLogs(changes []repository.ChangeLog) (*File, error) {
	fileName := changeLogsFileName(time.Now())
	file, err := s.createFileWithJson(fileName, changes)

	if err != nil {
		slog.Error("GoogleDriveService[UploadChangeLogs] Create file", "fileName", fileName, "error", err)
		return nil, err
	}

	return s.createFileFromGoogleDrive(file), nil
}

func (s *GoogleDriveService) GetFileContent(fileId string) ([]byte, error) {
	return s.geFileContent(fileId)
}

func (s *GoogleDriveService) PruneOldChanges(timestamp time.Time) error {
	resp, err := s.files().Q(fmt.Sprintf("createdTime < '%s'", s.formatTime(timestamp))).Do()
	if err != nil {
		slog.Error("GoogleDriveService[PruneOldChanges] Get list CHANGES_FILE_PREFIX", "error", err)
		return err
	}

	var wg sync.WaitGroup
	for _, file := range resp.Files {
		wg.Add(1)
		go func(file *drive.File) {
			defer wg.Done()
			err := s.service.Files.Delete(file.Id).Do()
			if err != nil {
				slog.Error("GoogleDriveService[PruneOldChanges] Unable to delete file", "fileName", file.Name, "error", err)
			} else {
				slog.Debug("GoogleDriveService[PruneOldChanges] File deleted", "fileName", file.Name)
			}
		}(file)
	}

	wg.Wait()

	return nil
}
