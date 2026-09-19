// Copyright (c) 2024
// Licensed under the MIT License. See LICENSE file in the root directory.

package google

import (
	"context"
	"fmt"
	"log/slog"
	"myscript/internal/repository"
	"net/http"
	"net/url"
	"strings"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	oauth2v2 "google.golang.org/api/oauth2/v2"
	"google.golang.org/api/option"
)

const revokeURL = "https://oauth2.googleapis.com/revoke"

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

func (c *GoogleClient) GetSavedToken() *repository.GoogleAuthToken {
	return c.repository.GetGoogleAuthToken()
}

// GetClientFromSavedToken refreshes the saved token when it is due and returns
// a client that keeps refreshing, saving each new token as it goes.
func (c *GoogleClient) GetClientFromSavedToken() (*http.Client, error) {
	config, err := c.getConfig()
	if err != nil {
		return nil, err
	}

	saved := c.repository.GetGoogleAuthToken()
	if saved == nil {
		return nil, ErrNoToken
	}

	previous := saved.AuthToken.Data()

	refreshed, err := config.TokenSource(context.Background(), previous).Token()
	if err != nil {
		return nil, err
	}

	refreshed = keepRefreshToken(previous, refreshed)
	if previous == nil || refreshed.AccessToken != previous.AccessToken {
		c.repository.UpdateGoogleAuthToken(refreshed)
	}

	source := &persistingTokenSource{
		source:     config.TokenSource(context.Background(), refreshed),
		repository: c.repository,
		previous:   refreshed,
	}

	return oauth2.NewClient(context.Background(), source), nil
}

// Revoke tells Google to forget the grant, so disconnecting here also
// disconnects on the account's side rather than leaving a live token behind.
func (c *GoogleClient) Revoke() error {
	saved := c.repository.GetGoogleAuthToken()
	if saved == nil {
		return nil
	}

	token := saved.AuthToken.Data()
	if token == nil {
		return nil
	}

	value := token.RefreshToken
	if value == "" {
		value = token.AccessToken
	}
	if value == "" {
		return nil
	}

	response, err := http.PostForm(revokeURL, url.Values{"token": {value}})
	if err != nil {
		return err
	}
	defer response.Body.Close()

	// 400 means Google already considers it gone, which is the outcome we wanted.
	if response.StatusCode >= 300 && response.StatusCode != http.StatusBadRequest {
		return fmt.Errorf("revoking the Google token returned %s", response.Status)
	}
	return nil
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

	if err := c.verifyScopes(token); err != nil {
		slog.Error("Unable to verify scopes", "error", err)
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

func (c *GoogleClient) verifyScopes(token *oauth2.Token) error {
	rawToken, ok := token.Extra("scope").(string)
	if !ok {
		slog.Error("Unable to extract scopes from token")
	}

	// Check if the token has all required scopes
	grantedScopes := strings.Fields(rawToken)
	grantedSet := make(map[string]struct{})
	for _, s := range grantedScopes {
		grantedSet[s] = struct{}{}
	}

	var missingScopes []string
	for _, required := range SCOPES {
		if _, exists := grantedSet[required]; !exists {
			missingScopes = append(missingScopes, required)
		}
	}

	if len(missingScopes) > 0 {
		slog.Error("Token is missing required scopes", "missing", missingScopes)
		return fmt.Errorf("authorization is incomplete due to missing scopes")
	}

	return nil
}

func (c *GoogleClient) getConfig() (*oauth2.Config, error) {
	if c.config == nil {
		return nil, ErrNoCredentials
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
