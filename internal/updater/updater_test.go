package updater

import (
	"archive/zip"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeBundleZip(t *testing.T, path string) {
	t.Helper()

	file, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()

	archive := zip.NewWriter(file)
	add := func(name, body string, mode os.FileMode) {
		header := &zip.FileHeader{Name: name, Method: zip.Deflate}
		header.SetMode(mode)
		w, err := archive.CreateHeader(header)
		if err != nil {
			t.Fatal(err)
		}
		w.Write([]byte(body))
	}

	// Ordered the way ditto writes a bundle: Info.plist lands before the binary.
	add("MyScript.app/Contents/Info.plist", "<plist/>", 0o644)
	add("MyScript.app/Contents/MacOS/MyScript", "#!/bin/sh\necho v2\n", 0o755)
	add("MyScript.app/Contents/Frameworks/Current", "A", os.ModeSymlink|0o777)
	// ditto --sequesterRsrc writes resource forks into a sibling tree.
	add("__MACOSX/MyScript.app/Contents/._Info.plist", "rsrc", 0o644)

	if err := archive.Close(); err != nil {
		t.Fatal(err)
	}
}

// Installing the first zip entry over the running executable would destroy a
// macOS install.
func TestUnzipBundleRebuildsWholeBundle(t *testing.T) {
	dir := t.TempDir()
	archive := filepath.Join(dir, "update.zip")
	writeBundleZip(t, archive)

	bundle := filepath.Join(dir, "MyScript.app")
	if err := unzipBundle(archive, bundle); err != nil {
		t.Fatal(err)
	}

	binary := filepath.Join(bundle, "Contents", "MacOS", "MyScript")
	body, err := os.ReadFile(binary)
	if err != nil {
		t.Fatalf("the executable is missing from the installed bundle: %v", err)
	}
	if !strings.Contains(string(body), "echo v2") {
		t.Fatalf("the executable holds the wrong content: %q", body)
	}

	info, err := os.Stat(binary)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm()&0o111 == 0 {
		t.Fatalf("the executable lost its executable bit: %v", info.Mode())
	}

	if _, err := os.Stat(filepath.Join(bundle, "Contents", "Info.plist")); err != nil {
		t.Fatalf("Info.plist is missing: %v", err)
	}

	link, err := os.Lstat(filepath.Join(bundle, "Contents", "Frameworks", "Current"))
	if err != nil {
		t.Fatal(err)
	}
	if link.Mode()&os.ModeSymlink == 0 {
		t.Fatal("a symlink in the bundle was replaced by a regular file")
	}
}

func TestUnzipBundleRejectsPathEscape(t *testing.T) {
	dir := t.TempDir()
	archive := filepath.Join(dir, "evil.zip")

	file, err := os.Create(archive)
	if err != nil {
		t.Fatal(err)
	}
	writer := zip.NewWriter(file)
	w, _ := writer.Create("MyScript.app/../../../../etc/escaped")
	w.Write([]byte("owned"))
	writer.Close()
	file.Close()

	if err := unzipBundle(archive, filepath.Join(dir, "out")); err == nil {
		t.Fatal("an entry escaping the destination was extracted")
	}
}

func TestParseChecksums(t *testing.T) {
	sum := strings.Repeat("ab", 32)
	body := strings.Join([]string{
		sum + "  myscript-darwin-arm64.zip",
		strings.Repeat("cd", 32) + " *myscript-linux-amd64.tar.gz",
		"garbage line",
		"",
	}, "\n")

	sums, err := parseChecksums(strings.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	if len(sums) != 2 {
		t.Fatalf("expected 2 checksums, got %d", len(sums))
	}

	want, _ := hex.DecodeString(sum)
	if !bytes.Equal(sums["myscript-darwin-arm64.zip"], want) {
		t.Fatal("the darwin checksum did not round-trip")
	}
	if _, ok := sums["myscript-linux-amd64.tar.gz"]; !ok {
		t.Fatal("the binary marker prefix was not stripped")
	}
}

func TestVerifyChecksumRejectsTamperedDownload(t *testing.T) {
	path := filepath.Join(t.TempDir(), "asset")
	os.WriteFile(path, []byte("the real release"), 0o644)

	genuine := sha256.Sum256([]byte("the real release"))
	if err := verifyChecksum(path, genuine[:]); err != nil {
		t.Fatalf("a genuine download was rejected: %v", err)
	}

	os.WriteFile(path, []byte("swapped by an attacker"), 0o644)
	if err := verifyChecksum(path, genuine[:]); err == nil {
		t.Fatal("a tampered download passed verification")
	}
}

func TestAssetNameMatchesReleaseArtifacts(t *testing.T) {
	published := map[string]bool{
		"myscript-windows-amd64-installer.exe": true,
		"myscript-darwin-amd64.zip":            true,
		"myscript-darwin-arm64.zip":            true,
		"myscript-linux-amd64.tar.gz":          true,
		"myscript-linux-arm64.tar.gz":          true,
		"myscript-linux-amd64.AppImage":        true,
		"myscript-linux-arm64.AppImage":        true,
	}

	updater := &Updater{}
	if !published[updater.assetName()] {
		t.Fatalf("the updater asks for %q, which the release never publishes", updater.assetName())
	}
}

func TestUnzipBundleSkipsDittoMetadata(t *testing.T) {
	dir := t.TempDir()
	archive := filepath.Join(dir, "update.zip")
	writeBundleZip(t, archive)

	bundle := filepath.Join(dir, "MyScript.app")
	if err := unzipBundle(archive, bundle); err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(filepath.Join(dir, "__MACOSX")); err == nil {
		t.Fatal("ditto metadata was extracted alongside the bundle")
	}
	if _, err := os.Stat(filepath.Join(bundle, "__MACOSX")); err == nil {
		t.Fatal("ditto metadata was extracted into the bundle")
	}
}

// Apple silicon refuses to exec an arm64 binary whose bytes no longer match
// its ad-hoc signature, so extraction must reproduce the executable exactly.
func TestUnzipBundlePreservesBytesExactly(t *testing.T) {
	payload := make([]byte, 4096)
	for i := range payload {
		payload[i] = byte(i * 7 % 251)
	}

	dir := t.TempDir()
	archive := filepath.Join(dir, "update.zip")

	file, err := os.Create(archive)
	if err != nil {
		t.Fatal(err)
	}
	writer := zip.NewWriter(file)
	header := &zip.FileHeader{Name: "MyScript.app/Contents/MacOS/MyScript", Method: zip.Deflate}
	header.SetMode(0o755)
	entry, err := writer.CreateHeader(header)
	if err != nil {
		t.Fatal(err)
	}
	entry.Write(payload)
	writer.Close()
	file.Close()

	bundle := filepath.Join(dir, "MyScript.app")
	if err := unzipBundle(archive, bundle); err != nil {
		t.Fatal(err)
	}

	installed, err := os.ReadFile(filepath.Join(bundle, "Contents", "MacOS", "MyScript"))
	if err != nil {
		t.Fatal(err)
	}
	if sha256.Sum256(installed) != sha256.Sum256(payload) {
		t.Fatal("extraction altered the executable, which invalidates its code signature")
	}
}
