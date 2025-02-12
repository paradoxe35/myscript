// Copyright (c) 2024
// Licensed under the MIT License. See LICENSE file in the root directory.

package filesystem

import (
	"log"
	"os"
	"path/filepath"
)

const homeDirName = ".myscript"

var HOME_DIR string

func createDirectoryInHome(dirName string) string {
	// Get the user's home directory
	homeDir, err := os.UserHomeDir()

	if err != nil {
		log.Fatalf("error getting home directory: %s", err)
	}

	// Create the full path by joining home directory with the new directory name
	newDirPath := filepath.Join(homeDir, dirName)

	// Check if directory already exists
	if _, err := os.Stat(newDirPath); err == nil {
		return newDirPath
	} else if !os.IsNotExist(err) {
		log.Fatalf("error checking directory: %s", err)
	}

	// Create the directory with default permissions (0755)
	err = os.Mkdir(newDirPath, 0755)
	if err != nil {
		log.Fatalf("error creating directory: %s", err)
	}

	return newDirPath
}

func init() {
	HOME_DIR = createDirectoryInHome(homeDirName)
}
