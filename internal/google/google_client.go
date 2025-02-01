package google

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
)

var SCOPES = []string{
	"https://www.googleapis.com/auth/userinfo.email",
	"https://www.googleapis.com/auth/userinfo.profile",
	"https://www.googleapis.com/auth/drive.file",
	"https://www.googleapis.com/auth/drive.appdata",
}

type GoogleClient struct {
	repository *repository.GoogleAuthTokenRepository
	config     *oauth2.Config
}

func NewGoogleClient(credentials []byte, repository *repository.GoogleAuthTokenRepository) *GoogleClient {
	if len(credentials) == 0 {
		return &GoogleClient{}
	}

	config, err := google.ConfigFromJSON(credentials, SCOPES...)

	if err != nil {
		return &GoogleClient{repository: repository}
	}

	return &GoogleClient{
		repository: repository,
		config:     config,
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
	token := c.repository.GetGoogleAuthToken()
	if token == nil {
		return nil, fmt.Errorf("invalid Google credentials, no token found")
	}

	authToken := token.AuthToken.Data()

	tokenSource, err := c.config.TokenSource(context.Background(), token.AuthToken.Data()).Token()
	if err != nil {
		return nil, err
	}

	if tokenSource.RefreshToken != authToken.RefreshToken {
		c.repository.UpdateGoogleAuthToken(tokenSource)
	}

	client, err := c.GetClient(tokenSource)
	if err != nil {
		return nil, err
	}

	return client, nil
}

func (c *GoogleClient) GenerateAuthURL(redirectURI string) (string, error) {
	if _, err := c.getConfig(); err != nil {
		return "", err
	}

	authURL := c.config.AuthCodeURL(
		"state-token",
		oauth2.AccessTypeOffline,
		oauth2.SetAuthURLParam("redirect_uri", redirectURI),
	)

	slog.Debug(
		"Go to the following link in your browser then type the",
		"authorization code", authURL,
	)

	return authURL, nil
}

func (c *GoogleClient) SaveAuthToken(code, redirectURI string) (*oauth2.Token, error) {
	if _, err := c.getConfig(); err != nil {
		return nil, err
	}

	token, err := c.config.Exchange(
		context.Background(), code,
		oauth2.AccessTypeOffline,
		oauth2.SetAuthURLParam("redirect_uri", redirectURI),
	)
	if err != nil {
		slog.Error("Unable to retrieve token from web", "error", err)
		return nil, err
	}

	userInfo, err := c.getUserInfo(token)
	if err != nil {
		slog.Error("Unable to retrieve user info", "error", err)
		return nil, err
	}

	c.repository.SaveGoogleAuthToken(userInfo, token)

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
