// Copyright (c) 2024
// Licensed under the MIT License. See LICENSE file in the root directory.

package google

import (
	"errors"
	"myscript/internal/repository"
	"strings"

	"golang.org/x/oauth2"
)

var (
	ErrNoCredentials = errors.New("this build has no Google credentials")
	ErrNoToken       = errors.New("not connected to Google Drive")
)

// Matched on text because the transport reports these as plain errors.
var authErrorMarkers = []string{
	"invalid_grant",
	"invalid_token",
	"token has been expired or revoked",
	"invalid credentials",
	"invalid authentication credentials",
	"unauthorized",
	"unauthenticated",
}

// A failure only signing in again can fix, unlike a network blip.
func IsAuthError(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, ErrNoToken) {
		return true
	}

	var retrieve *oauth2.RetrieveError
	if errors.As(err, &retrieve) {
		switch retrieve.Response.StatusCode {
		case 400, 401, 403:
			return true
		}
	}

	message := strings.ToLower(err.Error())
	for _, marker := range authErrorMarkers {
		if strings.Contains(message, marker) {
			return true
		}
	}
	return false
}

// Google returns the refresh token only on first consent, so carry it forward.
func keepRefreshToken(previous, refreshed *oauth2.Token) *oauth2.Token {
	if refreshed == nil {
		return previous
	}
	if refreshed.RefreshToken == "" && previous != nil {
		copied := *refreshed
		copied.RefreshToken = previous.RefreshToken
		return &copied
	}
	return refreshed
}

// Saves tokens the transport refreshes mid-session.
type persistingTokenSource struct {
	source     oauth2.TokenSource
	repository *repository.GoogleAuthTokenRepository
	previous   *oauth2.Token
}

func (p *persistingTokenSource) Token() (*oauth2.Token, error) {
	token, err := p.source.Token()
	if err != nil {
		return nil, err
	}

	token = keepRefreshToken(p.previous, token)
	if p.previous == nil || token.AccessToken != p.previous.AccessToken {
		p.previous = token
		p.repository.UpdateGoogleAuthToken(token)
	}

	return token, nil
}
