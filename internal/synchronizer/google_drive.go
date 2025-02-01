package synchronizer

import (
	"context"
	"net/http"

	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
)

type GoogleDriveService struct {
	ctx          context.Context
	service      *drive.Service
	serviceError error
}

func NewGoogleDriveService(client *http.Client) *GoogleDriveService {
	ctx := context.Background()

	service, err := drive.NewService(ctx, option.WithHTTPClient(client))

	return &GoogleDriveService{
		ctx:          ctx,
		service:      service,
		serviceError: err,
	}
}
