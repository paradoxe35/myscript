// Copyright (c) 2024
// Licensed under the MIT License. See LICENSE file in the root directory.

package utils

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"io"
	"log/slog"
	"os"
	"path/filepath"
)

type FileArchiver struct {
	globPattern string
}

func NewFileArchiver(globPattern string) *FileArchiver {
	return &FileArchiver{globPattern: globPattern}
}

func (a *FileArchiver) Archive() (*bytes.Buffer, error) {
	matches, err := filepath.Glob(a.globPattern)
	if err != nil {
		slog.Error("[FileArchiver] Error finding files", "error", err)
		return nil, err
	}

	// Step 2: Compress the files into a gzip archive
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gz)

	for _, file := range matches {
		err := a.addToArchive(tw, file)
		if err != nil {
			slog.Error("[FileArchiver] Error adding file to archive", "file", file, "error", err)
			return nil, err
		}
	}

	// Close the tar writer and gzip writer
	if err := tw.Close(); err != nil {
		slog.Error("[FileArchiver] Failed to close tar writer", "error", err)
		return nil, err
	}
	if err := gz.Close(); err != nil {
		slog.Error("[FileArchiver] Failed to close gzip writer", "error", err)
		return nil, err
	}

	return &buf, nil
}

func (a *FileArchiver) addToArchive(tw *tar.Writer, filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return err
	}

	header, err := tar.FileInfoHeader(info, info.Name())
	if err != nil {
		return err
	}

	header.Name = filepath.Base(filename)

	if err := tw.WriteHeader(header); err != nil {
		return err
	}

	_, err = io.Copy(tw, file)
	return err
}
