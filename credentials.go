package main

import (
	"bytes"
	"errors"
	"io/fs"
	"log/slog"
	"strings"
)

func readFile(file string) []byte {
	fileData, err := credentials.ReadFile("credentials/" + file)
	if err != nil {
		if !errors.Is(err, fs.ErrNotExist) {
			slog.Error("Unexpected error opening GitHub token", "error", err)
		}
		return nil
	}

	// Remove UTF-8 BOM if present
	data := bytes.TrimPrefix(fileData, []byte{0xEF, 0xBB, 0xBF})

	// Remove UTF-16 BOM if present
	data = bytes.TrimPrefix(data, []byte{0xFF, 0xFE})

	return data
}

func readGoogleCredentials() []byte {
	return readFile("google-credentials.json")
}

func readGitHubToken() string {
	data := readFile("github-token.txt")
	// Handle Windows/Unix line endings
	token := strings.TrimSpace(strings.ReplaceAll(string(data), "\r\n", "\n"))

	return token
}
