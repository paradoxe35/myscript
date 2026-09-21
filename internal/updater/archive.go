// Copyright (c) 2024
// Licensed under the MIT License. See LICENSE file in the root directory.

package updater

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// Drops the bundle's own directory so dest becomes the bundle. Symlinks are
// kept: copying a framework link's target doubles the size and breaks the signature.
func unzipBundle(archivePath, dest string) error {
	reader, err := zip.OpenReader(archivePath)
	if err != nil {
		return err
	}
	defer reader.Close()

	root := bundleRoot(reader.File)
	if root == "" {
		return fmt.Errorf("the archive holds no .app bundle")
	}

	extracted := 0
	for _, file := range reader.File {
		relative, inside := strings.CutPrefix(file.Name, root)
		if !inside || relative == "" {
			continue
		}

		path, err := safeJoin(dest, relative)
		if err != nil {
			return err
		}
		if err := extractZipEntry(file, path); err != nil {
			return err
		}
		extracted++
	}

	if extracted == 0 {
		return fmt.Errorf("the bundle in the archive is empty")
	}
	return nil
}

// bundleRoot is the "Something.app/" prefix every entry shares.
func bundleRoot(files []*zip.File) string {
	for _, file := range files {
		name := strings.TrimPrefix(file.Name, "./")
		if before, _, found := strings.Cut(name, ".app/"); found {
			return before + ".app/"
		}
	}
	return ""
}

func extractZipEntry(file *zip.File, path string) error {
	info := file.FileInfo()

	if info.IsDir() {
		return os.MkdirAll(path, 0o755)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}

	source, err := file.Open()
	if err != nil {
		return err
	}
	defer source.Close()

	if info.Mode()&os.ModeSymlink != 0 {
		target, err := io.ReadAll(source)
		if err != nil {
			return err
		}
		os.Remove(path)
		return os.Symlink(string(target), path)
	}

	return writeFile(path, source, info.Mode())
}

// The caller streams the executable straight into the installer, so a
// half-written copy is never on disk.
func openBinaryInTarGz(archivePath string) (io.Reader, io.Closer, error) {
	file, err := os.Open(archivePath)
	if err != nil {
		return nil, nil, err
	}

	gz, err := gzip.NewReader(file)
	if err != nil {
		file.Close()
		return nil, nil, err
	}

	closer := closerFunc(func() error {
		gz.Close()
		return file.Close()
	})

	entries := tar.NewReader(gz)
	for {
		header, err := entries.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			closer.Close()
			return nil, nil, err
		}
		if header.Typeflag == tar.TypeReg {
			return entries, closer, nil
		}
	}

	closer.Close()
	return nil, nil, fmt.Errorf("the archive holds no executable")
}

type closerFunc func() error

func (c closerFunc) Close() error { return c() }

func writeFile(path string, content io.Reader, mode os.FileMode) error {
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, mode)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = io.Copy(file, content)
	return err
}

// Refuses an entry whose name would escape the destination.
func safeJoin(dest, name string) (string, error) {
	path := filepath.Join(dest, filepath.FromSlash(name))
	if !strings.HasPrefix(path, filepath.Clean(dest)+string(os.PathSeparator)) {
		return "", fmt.Errorf("archive entry %q escapes the destination", name)
	}
	return path, nil
}
