// Copyright (c) 2024
// Licensed under the MIT License. See LICENSE file in the root directory.

package google

import (
	"errors"
	"fmt"
	"myscript/internal/repository"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	"golang.org/x/oauth2"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func newTokenRepository(t *testing.T) *repository.GoogleAuthTokenRepository {
	t.Helper()

	db, err := gorm.Open(
		sqlite.Open(filepath.Join(t.TempDir(), "test.sqlite")),
		&gorm.Config{Logger: logger.Discard},
	)
	if err != nil {
		t.Fatalf("open database: %v", err)
	}
	if err := db.AutoMigrate(&repository.GoogleAuthToken{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	return repository.NewGoogleAuthTokenRepository(db)
}

func TestIsAuthErrorRecognisesADeadGrant(t *testing.T) {
	cases := map[string]bool{
		"oauth2: cannot fetch token: 400 Bad Request invalid_grant": true,
		"Token has been expired or revoked.":                        true,
		"Request had invalid authentication credentials":            true,
		"unauthorized_client":                                       true,
		"dial tcp: lookup accounts.google.com: no such host":        false,
		"googleapi: Error 500: backend error":                       false,
		"context deadline exceeded":                                 false,
	}

	for message, want := range cases {
		if got := IsAuthError(errors.New(message)); got != want {
			t.Errorf("IsAuthError(%q) = %v, want %v", message, got, want)
		}
	}

	if IsAuthError(nil) {
		t.Error("no error is not an auth error")
	}
	if !IsAuthError(ErrNoToken) {
		t.Error("a missing token needs signing in again")
	}
}

func TestIsAuthErrorReadsTheStatus(t *testing.T) {
	unauthorized := &oauth2.RetrieveError{
		Response: &http.Response{StatusCode: http.StatusUnauthorized},
	}
	if !IsAuthError(fmt.Errorf("refresh: %w", unauthorized)) {
		t.Error("a 401 should be an auth error even when wrapped")
	}

	unavailable := &oauth2.RetrieveError{
		Response: &http.Response{StatusCode: http.StatusServiceUnavailable},
	}
	if IsAuthError(unavailable) {
		t.Error("a 503 is worth retrying, not re-authenticating")
	}
}

func TestRefreshedTokenKeepsTheRefreshToken(t *testing.T) {
	previous := &oauth2.Token{AccessToken: "old", RefreshToken: "keep-me"}

	// Google returns the refresh token only on first consent.
	refreshed := keepRefreshToken(previous, &oauth2.Token{AccessToken: "new"})

	if refreshed.RefreshToken != "keep-me" {
		t.Errorf("RefreshToken = %q, losing it means the account cannot refresh again", refreshed.RefreshToken)
	}
	if refreshed.AccessToken != "new" {
		t.Errorf("AccessToken = %q", refreshed.AccessToken)
	}

	// A response that does carry one wins.
	rotated := keepRefreshToken(previous, &oauth2.Token{AccessToken: "new", RefreshToken: "rotated"})
	if rotated.RefreshToken != "rotated" {
		t.Errorf("RefreshToken = %q, a new one should be adopted", rotated.RefreshToken)
	}
}

func TestSavedTokenIsRefreshedAndStored(t *testing.T) {
	var refreshes int

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		refreshes++
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"access_token":"fresh","token_type":"Bearer","expires_in":3600}`)
	}))
	defer server.Close()

	repo := newTokenRepository(t)
	repo.SaveGoogleAuthToken(nil, &oauth2.Token{
		AccessToken:  "stale",
		RefreshToken: "keep-me",
		Expiry:       time.Now().Add(-time.Hour),
	})

	client := &GoogleClient{
		repository: repo,
		config: &oauth2.Config{
			ClientID:     "id",
			ClientSecret: "secret",
			Endpoint:     oauth2.Endpoint{TokenURL: server.URL},
		},
	}

	if _, err := client.GetClientFromSavedToken(); err != nil {
		t.Fatalf("GetClientFromSavedToken: %v", err)
	}

	if refreshes != 1 {
		t.Errorf("made %d refresh calls, want 1", refreshes)
	}

	stored := repo.GetGoogleAuthToken().AuthToken.Data()
	if stored.AccessToken != "fresh" {
		t.Errorf("stored AccessToken = %q, the refreshed token should be saved", stored.AccessToken)
	}
	if stored.RefreshToken != "keep-me" {
		t.Errorf("stored RefreshToken = %q, it must survive a refresh", stored.RefreshToken)
	}
}

func TestAValidTokenIsNotRefreshed(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("a token that has not expired should not be refreshed")
	}))
	defer server.Close()

	repo := newTokenRepository(t)
	repo.SaveGoogleAuthToken(nil, &oauth2.Token{
		AccessToken:  "current",
		RefreshToken: "keep-me",
		Expiry:       time.Now().Add(time.Hour),
	})

	client := &GoogleClient{
		repository: repo,
		config: &oauth2.Config{
			ClientID: "id", ClientSecret: "secret",
			Endpoint: oauth2.Endpoint{TokenURL: server.URL},
		},
	}

	if _, err := client.GetClientFromSavedToken(); err != nil {
		t.Fatal(err)
	}
}

func TestARevokedGrantSurfacesAsAnAuthError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, `{"error":"invalid_grant","error_description":"Token has been expired or revoked."}`)
	}))
	defer server.Close()

	repo := newTokenRepository(t)
	repo.SaveGoogleAuthToken(nil, &oauth2.Token{
		AccessToken:  "stale",
		RefreshToken: "revoked",
		Expiry:       time.Now().Add(-time.Hour),
	})

	client := &GoogleClient{
		repository: repo,
		config: &oauth2.Config{
			ClientID: "id", ClientSecret: "secret",
			Endpoint: oauth2.Endpoint{TokenURL: server.URL},
		},
	}

	_, err := client.GetClientFromSavedToken()
	if err == nil {
		t.Fatal("expected the refresh to fail")
	}
	if !IsAuthError(err) {
		t.Errorf("error %q should be recognised as needing a new sign-in", err)
	}
}

func TestWithoutCredentialsNothingPanics(t *testing.T) {
	client := &GoogleClient{repository: newTokenRepository(t)}

	if _, err := client.GetClientFromSavedToken(); !errors.Is(err, ErrNoCredentials) {
		t.Fatalf("got %v, want ErrNoCredentials", err)
	}
}

func TestWithoutATokenTheErrorSaysSo(t *testing.T) {
	client := &GoogleClient{
		repository: newTokenRepository(t),
		config:     &oauth2.Config{ClientID: "id"},
	}

	if _, err := client.GetClientFromSavedToken(); !errors.Is(err, ErrNoToken) {
		t.Fatalf("got %v, want ErrNoToken", err)
	}
}

func TestAMidSessionRefreshIsSaved(t *testing.T) {
	repo := newTokenRepository(t)
	repo.SaveGoogleAuthToken(nil, &oauth2.Token{AccessToken: "first", RefreshToken: "keep-me"})

	source := &persistingTokenSource{
		source:     oauth2.StaticTokenSource(&oauth2.Token{AccessToken: "second"}),
		repository: repo,
		previous:   &oauth2.Token{AccessToken: "first", RefreshToken: "keep-me"},
	}

	if _, err := source.Token(); err != nil {
		t.Fatal(err)
	}

	stored := repo.GetGoogleAuthToken().AuthToken.Data()
	if stored.AccessToken != "second" {
		t.Errorf("stored AccessToken = %q, a refresh during a sync should be saved", stored.AccessToken)
	}
	if stored.RefreshToken != "keep-me" {
		t.Errorf("stored RefreshToken = %q", stored.RefreshToken)
	}
}
