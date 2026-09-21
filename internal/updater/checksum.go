// Copyright (c) 2024
// Licensed under the MIT License. See LICENSE file in the root directory.

package updater

import (
	"bufio"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"strings"
)

const checksumsAsset = "checksums.txt"

// Reads sha256sum output. Names may carry the "*" binary marker, and the
// release flattens directories, so only the base name is matched.
func parseChecksums(r io.Reader) (map[string][]byte, error) {
	sums := make(map[string][]byte)

	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		digest, name, found := strings.Cut(strings.TrimSpace(scanner.Text()), " ")
		if !found {
			continue
		}

		name = strings.TrimPrefix(strings.TrimSpace(name), "*")
		if name == "" {
			continue
		}

		sum, err := hex.DecodeString(digest)
		if err != nil || len(sum) != sha256.Size {
			continue
		}
		sums[name] = sum
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}
	if len(sums) == 0 {
		return nil, fmt.Errorf("%s lists no checksums", checksumsAsset)
	}
	return sums, nil
}

func verifyChecksum(path string, expected []byte) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	digest := sha256.New()
	if _, err := io.Copy(digest, file); err != nil {
		return err
	}

	actual := digest.Sum(nil)
	if subtle.ConstantTimeCompare(actual, expected) != 1 {
		return fmt.Errorf(
			"the download is corrupt or has been tampered with: expected sha256 %s, got %s",
			hex.EncodeToString(expected), hex.EncodeToString(actual),
		)
	}
	return nil
}
