// Copyright (c) 2024
// Licensed under the MIT License. See LICENSE file in the root directory.

package synchronizer

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"myscript/internal/database"
	"os"
	"path/filepath"
)

func DecompressSnapshotFile(data []byte, tmpDir string) (string, error) {
	gzReader, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return "", fmt.Errorf("failed to create gzip reader: %v", err)
	}
	defer gzReader.Close()

	tarReader := tar.NewReader(gzReader)

	var dbPath string
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", fmt.Errorf("failed to read tar entry: %v", err)
		}

		target := filepath.Join(tmpDir, header.Name)
		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, os.ModePerm); err != nil {
				return "", fmt.Errorf("failed to create directory: %v", err)
			}
		case tar.TypeReg:
			outFile, err := os.Create(target)
			if err != nil {
				return "", fmt.Errorf("failed to create file: %v", err)
			}
			defer outFile.Close()

			if _, err := io.Copy(outFile, tarReader); err != nil {
				return "", fmt.Errorf("failed to write file: %v", err)
			}

			// Check if the file is the SQLite database
			if filepath.Base(header.Name) == database.DB_BASE_NAME {
				dbPath = target
			}
		}
	}

	if dbPath == "" {
		return "", fmt.Errorf("%s not found in the archive", database.DB_BASE_NAME)
	}

	return dbPath, nil
}
