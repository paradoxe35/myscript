// Copyright (c) 2024
// Licensed under the MIT License. See LICENSE file in the root directory.

package main

import (
	"bytes"
	"errors"
	"io/fs"
	"log/slog"
)

func readFile(file string) []byte {
	fileData, err := credentials.ReadFile("credentials/" + file)
	if err != nil {
		// A missing file is the normal case for a build without credentials.
		if !errors.Is(err, fs.ErrNotExist) {
			slog.Error("Unexpected error reading an embedded credential", "file", file, "error", err)
		}
		return nil
	}

	// Strip a UTF-8 or UTF-16 BOM
	data := bytes.TrimPrefix(fileData, []byte{0xEF, 0xBB, 0xBF})
	data = bytes.TrimPrefix(data, []byte{0xFF, 0xFE})

	return data
}

func readGoogleCredentials() []byte {
	return readFile("google-credentials.json")
}
