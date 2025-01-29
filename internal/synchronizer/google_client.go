package synchronizer

import (
	"context"
	"fmt"
	"log/slog"
	"myscript/internal/repository"
	"net/http"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	oauth2v2 "google.golang.org/api/oauth2/v2"
	"google.golang.org/api/option"
	"gorm.io/gorm"
)

var SCOPES = []string{
	"https://www.googleapis.com/auth/userinfo.email",
	"https://www.googleapis.com/auth/userinfo.profile",
	"https://www.googleapis.com/auth/drive.file",
	"https://www.googleapis.com/auth/drive.appdata",
}

type GoogleClient struct {
	db     *gorm.DB
	config *oauth2.Config
}

func NewGoogleClient(credentials []byte, db *gorm.DB) *GoogleClient {
	config, err := google.ConfigFromJSON(credentials, SCOPES...)

	if err != nil {
		return &GoogleClient{}
	}

	return &GoogleClient{
		config: config,
	}
}

func (c *GoogleClient) HasCredentials() bool {
	_, err := c.getConfig()
	return err == nil
}

func (c *GoogleClient) GetClient(token *oauth2.Token) (*http.Client, error) {
	if _, err := c.getConfig(); err != nil {
		return nil, err
	}

	return c.config.Client(context.Background(), token), nil
}

func (c *GoogleClient) GetClientFromSavedToken() (*http.Client, error) {
	token := repository.GetGoogleAuthToken(c.db)

	if token == nil {
		return nil, fmt.Errorf("invalid Google credentials, no token found")
	}

	client, err := c.GetClient(token.AuthToken.Data())
	if err != nil {
		return nil, err
	}

	return client, nil
}

func (c *GoogleClient) GenerateAuthURL() (string, error) {
	if _, err := c.getConfig(); err != nil {
		return "", err
	}

	authURL := c.config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)

	slog.Debug(
		"Go to the following link in your browser then type the",
		"authorization code", authURL,
	)

	return authURL, nil
}

func (c *GoogleClient) SaveAuthToken(code string) (*oauth2.Token, error) {
	if _, err := c.getConfig(); err != nil {
		return nil, err
	}

	token, err := c.config.Exchange(context.Background(), code)
	if err != nil {
		return nil, err
	}

	userInfo, err := c.getUserInfo(token)
	if err != nil {
		return nil, err
	}

	repository.SaveGoogleAuthToken(c.db, userInfo, token)

	slog.Debug("Authentication successful and token saved", "user", userInfo.Email)

	return token, nil
}

func (c *GoogleClient) getConfig() (*oauth2.Config, error) {
	if c.config == nil {
		return nil, fmt.Errorf("invalid Google credentials, no config found")
	}
	return c.config, nil
}

func (c *GoogleClient) getUserInfo(token *oauth2.Token) (*oauth2v2.Userinfo, error) {
	ctx := context.Background()

	client, err := c.GetClient(token)
	if err != nil {
		return nil, err
	}

	oauth2Service, err := oauth2v2.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil, err
	}

	userInfo, err := oauth2Service.Userinfo.Get().Do()
	if err != nil {
		return nil, err
	}

	return userInfo, nil
}
