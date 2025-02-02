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

func (s *GoogleDriveService) formatTime(t time.Time) string {
	return t.Format(time.RFC3339)
}

func (s *GoogleDriveService) parseTime(t string) time.Time {
	parsedTime, _ := time.Parse(time.RFC3339, t)
	return parsedTime
}

func (s *GoogleDriveService) defaultTime() string {
	return INITIAL_SYNC_TIME.Format(time.RFC3339)
}

func (s *GoogleDriveService) files() *drive.FilesListCall {
	return s.service.Files.List().
		Spaces(PARENT_FOLDER)
}

func (s *GoogleDriveService) SortFilesAsc(files []*drive.File) []*drive.File {
	sort.Slice(files, func(i, j int) bool {
		timeI, _ := time.Parse(time.RFC3339, files[i].ModifiedTime)
		timeJ, _ := time.Parse(time.RFC3339, files[j].ModifiedTime)
		return timeI.Before(timeJ)
	})

	return files
}

func (s *GoogleDriveService) SortFilesDesc(files []*drive.File) []*drive.File {
	sort.Slice(files, func(i, j int) bool {
		timeI, _ := time.Parse(time.RFC3339, files[i].ModifiedTime)
		timeJ, _ := time.Parse(time.RFC3339, files[j].ModifiedTime)
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

func (s *GoogleDriveService) updateFileWithJson(file *drive.File, content interface{}) error {
	data, err := json.Marshal(content)
	if err != nil {
		slog.Error("GoogleDriveService[updateFileWithJson] Marshal content to JSON", "error", err)
		return err
	}

	if _, err := s.service.Files.Update(file.Id, file).Media(bytes.NewReader(data)).Do(); err != nil {
		slog.Error("GoogleDriveService[updateFileWithJson] Update file", "error", err)
		return err
	}

	return nil
}

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

func (s *GoogleDriveService) InitDevice(deviceID string) error {
	files, err := s.files().Q(fmt.Sprintf("name='%s'", DEVICES_SYNC_STATE_FILE)).Do()
	if err != nil {
		slog.Error("GoogleDriveService[InitDevice] Get list DEVICES_SYNC_STATE_FILE", "error", err)
		return err
	}

	state := DeviceSyncState{}

	if len(files.Files) > 0 {
		file := files.Files[0]

		fileContent, err := s.geFileContent(file.Id)
		if err != nil {
			slog.Error("GoogleDriveService[InitDevice] Get list DEVICES_SYNC_STATE_FILE content", "error", err)
			return err
		}

		if err = json.Unmarshal(fileContent, &state); err != nil {
			slog.Error("GoogleDriveService[InitDevice] Get list DEVICES_SYNC_STATE_FILE Unmarshal parses the JSON-encoded data", "error", err)
			return err
		}

		// Check if the device is already synced
		if _, ok := state[deviceID]; ok {
			return nil
		}

		state[deviceID] = &DeviceSyncStateValue{
			SyncTimeOffset: s.defaultTime(),
		}

		if err := s.updateFileWithJson(file, state); err != nil {
			slog.Error("GoogleDriveService[InitDevice] Update DEVICES_SYNC_STATE_FILE", "error", err)
			return err
		}
	} else {
		state[deviceID] = &DeviceSyncStateValue{
			SyncTimeOffset: s.defaultTime(),
		}

		if _, err := s.createFileWithJson(DEVICES_SYNC_STATE_FILE, state); err != nil {
			slog.Error("GoogleDriveService[InitDevice] Create DEVICES_SYNC_STATE_FILE", "error", err)
			return err
		}
	}

	return nil
}

func (s *GoogleDriveService) UpdateDeviceSyncTimeOffset(deviceID string, timeOffset time.Time) error {
	files, err := s.files().Q(fmt.Sprintf("name='%s'", DEVICES_SYNC_STATE_FILE)).Do()
	if err != nil {
		slog.Error("GoogleDriveService[UpdateDeviceSyncTimeOffset] Get list DEVICES_SYNC_STATE_FILE", "error", err)
		return err
	}

	if len(files.Files) > 0 {
		file := files.Files[0]

		fileContent, err := s.geFileContent(file.Id)
		if err != nil {
			slog.Error("GoogleDriveService[UpdateDeviceSyncTimeOffset] Get list DEVICES_SYNC_STATE_FILE content", "error", err)
			return err
		}

		state := DeviceSyncState{}
		if err = json.Unmarshal(fileContent, &state); err != nil {
			slog.Error("GoogleDriveService[UpdateDeviceSyncTimeOffset] Get list DEVICES_SYNC_STATE_FILE Unmarshal parses the JSON-encoded data", "error", err)
			return err
		}

		// Check if the device is already synced
		if _, ok := state[deviceID]; !ok {
			return fmt.Errorf("device %s is not synced", deviceID)
		}

		state[deviceID] = &DeviceSyncStateValue{
			SyncTimeOffset: s.formatTime(timeOffset),
		}

		if err := s.updateFileWithJson(file, state); err != nil {
			slog.Error("GoogleDriveService[UpdateDeviceSyncTimeOffset] Update DEVICES_SYNC_STATE_FILE", "error", err)
			return err
		}
	} else {
		return fmt.Errorf("device %s is not synced", deviceID)
	}

	return nil
}

func (s *GoogleDriveService) GetLatestDBSnapshot() (*File, error) {
	resp, err := s.files().Q(fmt.Sprintf("name contains '%s'", DB_SNAPSHOT_PREFIX)).Do()
	if err != nil {
		slog.Error("GoogleDriveService[GetLatestDBSnapshot] Get list DB_SNAPSHOT_PREFIX", "error", err)
		return nil, err
	}

	if len(resp.Files) == 0 {
		slog.Error("GoogleDriveService[GetLatestDBSnapshot] No DB snapshot found")
		return nil, fmt.Errorf("no DB snapshot found")
	}

	resp.Files = s.SortFilesDesc(resp.Files)

	file := resp.Files[0]

	return &File{
		ID:         file.Id,
		Name:       file.Name,
		IsSnapshot: true,
		ModifiedAt: s.parseTime(file.ModifiedTime),
	}, nil
}

func (s *GoogleDriveService) SaveDBSnapshot(content io.ReadSeeker) (*File, error) {
	fileName := snapshotFileName(time.Now())
	file, err := s.createRawFile(fileName, content)

	if err != nil {
		slog.Error("GoogleDriveService[SaveDBSnapshot] Create file", "fileName", fileName, "error", err)
		return nil, err
	}

	return &File{
		ID:         file.Id,
		Name:       file.Name,
		ModifiedAt: s.parseTime(file.ModifiedTime),
	}, nil
}

func (s *GoogleDriveService) GetFilesAfterTimeOffset(timeOffset time.Time) ([]File, error) {
	resp, err := s.files().Q(fmt.Sprintf("modifiedTime > '%s'", s.formatTime(timeOffset))).Do()

	if err != nil {
		slog.Error("GoogleDriveService[GetFilesAfterTimeOffset] Get list CHANGES_FILE_PREFIX", "error", err)
		return nil, err
	}

	resp.Files = s.SortFilesAsc(resp.Files)

	files := make([]File, len(resp.Files))
	for i, file := range resp.Files {
		files[i] = File{
			ID:         file.Id,
			Name:       file.Name,
			IsSnapshot: strings.HasPrefix(file.Name, DB_SNAPSHOT_PREFIX),
			ModifiedAt: s.parseTime(file.ModifiedTime),
		}
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

	return &File{
		ID:         file.Id,
		Name:       file.Name,
		ModifiedAt: s.parseTime(file.ModifiedTime),
	}, nil
}

func (s *GoogleDriveService) GetFileContent(fileId string) ([]byte, error) {
	return s.geFileContent(fileId)
}

func (s *GoogleDriveService) PruneOldChanges(timestamp time.Time) error {
	resp, err := s.files().Q(fmt.Sprintf("modifiedTime < '%s'", s.formatTime(timestamp))).Do()
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
