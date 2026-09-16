package stt

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"
)

func payload(size int) []byte {
	body := make([]byte, size)
	for i := range body {
		body[i] = byte(i % 251)
	}
	return body
}

func digest(body []byte) string {
	sum := sha256.Sum256(body)
	return hex.EncodeToString(sum[:])
}

// serve honours Range so resume can be exercised; ignoreRange reproduces a
// server that always answers 200.
func serve(t *testing.T, body []byte, ignoreRange bool) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := 0
		if spec := r.Header.Get("Range"); spec != "" && !ignoreRange {
			fmt.Sscanf(spec, "bytes=%d-", &start)
			if start >= len(body) {
				w.WriteHeader(http.StatusRequestedRangeNotSatisfiable)
				return
			}
			w.Header().Set("Content-Range",
				fmt.Sprintf("bytes %d-%d/%d", start, len(body)-1, len(body)))
			w.Header().Set("Content-Length", strconv.Itoa(len(body)-start))
			w.WriteHeader(http.StatusPartialContent)
			w.Write(body[start:])
			return
		}
		w.Header().Set("Content-Length", strconv.Itoa(len(body)))
		w.Write(body)
	}))
}

func testStore(t *testing.T, server *httptest.Server) *Store {
	t.Helper()
	store := NewStore(t.TempDir())
	if server != nil {
		store.urlFor = func(Model) string { return server.URL }
	}
	return store
}

func testModel(server *httptest.Server, body []byte) Model {
	return Model{
		ID:        "test/model",
		Name:      "Test Model",
		Repo:      strings.TrimPrefix(server.URL, "http://"),
		Filename:  "model.gguf",
		SizeBytes: int64(len(body)),
		SHA256:    digest(body),
	}
}

func TestDownloadVerifiesAndStores(t *testing.T) {
	body := payload(64 * 1024)
	server := serve(t, body, false)
	defer server.Close()

	store := testStore(t, server)
	model := testModel(server, body)

	if err := store.Download(context.Background(), model, nil); err != nil {
		t.Fatalf("Download: %v", err)
	}
	if !store.Downloaded(model) {
		t.Fatal("model should be reported as downloaded")
	}

	written, err := os.ReadFile(store.Path(model))
	if err != nil {
		t.Fatal(err)
	}
	if digest(written) != model.SHA256 {
		t.Fatal("stored file does not match the published checksum")
	}
	if _, err := os.Stat(store.Path(model) + partialSuffix); !os.IsNotExist(err) {
		t.Fatal("partial file should be gone after a successful download")
	}
}

func TestDownloadResumesFromPartial(t *testing.T) {
	body := payload(64 * 1024)
	server := serve(t, body, false)
	defer server.Close()

	store := testStore(t, server)
	model := testModel(server, body)

	half := len(body) / 2
	if err := os.WriteFile(store.Path(model)+partialSuffix, body[:half], 0o644); err != nil {
		t.Fatal(err)
	}

	var firstReport int64 = -1
	err := store.Download(context.Background(), model, func(p Progress) {
		if firstReport < 0 && p.Stage == StageDownloading {
			firstReport = p.Downloaded
		}
	})
	if err != nil {
		t.Fatalf("Download: %v", err)
	}
	if firstReport >= 0 && firstReport < int64(half) {
		t.Errorf("resumed from %d, expected to continue past %d", firstReport, half)
	}

	written, _ := os.ReadFile(store.Path(model))
	if digest(written) != model.SHA256 {
		t.Fatal("resumed file is corrupt")
	}
}

// A server that ignores Range replies 200 from byte zero; appending would duplicate the prefix.
func TestDownloadHandlesIgnoredRange(t *testing.T) {
	body := payload(32 * 1024)
	server := serve(t, body, true)
	defer server.Close()

	store := testStore(t, server)
	model := testModel(server, body)

	if err := os.WriteFile(store.Path(model)+partialSuffix, body[:1024], 0o644); err != nil {
		t.Fatal(err)
	}
	if err := store.Download(context.Background(), model, nil); err != nil {
		t.Fatalf("Download: %v", err)
	}

	written, _ := os.ReadFile(store.Path(model))
	if len(written) != len(body) || digest(written) != model.SHA256 {
		t.Fatal("file corrupted by an ignored Range header")
	}
}

func TestDownloadRejectsChecksumMismatch(t *testing.T) {
	body := payload(8 * 1024)
	server := serve(t, body, false)
	defer server.Close()

	store := testStore(t, server)
	model := testModel(server, body)
	model.SHA256 = digest(payload(16))

	err := store.Download(context.Background(), model, nil)
	if err == nil || !strings.Contains(err.Error(), "checksum") {
		t.Fatalf("a mismatched checksum must fail, got %v", err)
	}
	if _, err := os.Stat(store.Path(model) + partialSuffix); !os.IsNotExist(err) {
		t.Error("a corrupt partial should be deleted, not left to resume")
	}
	if store.Downloaded(model) {
		t.Error("a failed download must not be reported as present")
	}
}

// A full-size partial needs verifying, not another request: asking from EOF returns 416 and loops.
func TestDownloadVerifiesCompletePartial(t *testing.T) {
	body := payload(4096)
	server := serve(t, body, false)
	defer server.Close()

	store := testStore(t, server)
	model := testModel(server, body)

	if err := os.WriteFile(store.Path(model)+partialSuffix, body, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := store.Download(context.Background(), model, nil); err != nil {
		t.Fatalf("a complete partial should verify and move: %v", err)
	}
	if !store.Downloaded(model) {
		t.Fatal("model should be present")
	}
}

func TestDownloadStopsAtDeclaredSize(t *testing.T) {
	body := payload(8192)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write(append(body, payload(4096)...))
	}))
	defer server.Close()

	store := testStore(t, server)
	model := testModel(server, body)

	if err := store.Download(context.Background(), model, nil); err != nil {
		t.Fatalf("Download: %v", err)
	}
	info, err := os.Stat(store.Path(model))
	if err != nil {
		t.Fatal(err)
	}
	if info.Size() != model.SizeBytes {
		t.Fatalf("wrote %d bytes, expected to stop at %d", info.Size(), model.SizeBytes)
	}
}

func TestCancelKeepsThePartialAndReportsNotDownloading(t *testing.T) {
	body := payload(1 << 20)
	release := make(chan struct{})
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", strconv.Itoa(len(body)))
		w.Write(body[:4096])
		w.(http.Flusher).Flush()
		<-release
	}))
	defer server.Close()
	defer close(release)

	store := testStore(t, server)
	model := testModel(server, body)

	done := make(chan error, 1)
	go func() { done <- store.Download(context.Background(), model, nil) }()
	partial := store.Path(model) + partialSuffix
	for {
		if info, err := os.Stat(partial); err == nil && info.Size() > 0 {
			break
		}
		time.Sleep(time.Millisecond)
	}
	store.CancelDownload(model)

	if err := <-done; err == nil {
		t.Fatal("a cancelled download must not succeed")
	}
	if store.Downloading(model) {
		t.Error("cancelled download still reported in flight")
	}
	if _, err := os.Stat(partial); err != nil {
		t.Error("the partial should be kept to resume from")
	}
}

func TestDownloadGivesUpWhenTheServerStalls(t *testing.T) {
	body := payload(1 << 20)
	release := make(chan struct{})
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", strconv.Itoa(len(body)))
		w.Write(body[:4096])
		w.(http.Flusher).Flush()
		<-release
	}))
	defer server.Close()
	defer close(release)

	store := testStore(t, server)
	store.stall = 100 * time.Millisecond
	model := testModel(server, body)

	err := store.Download(context.Background(), model, nil)
	if err == nil || !strings.Contains(err.Error(), "stalled") {
		t.Fatalf("Download = %v, want the stall reported", err)
	}
	if _, err := os.Stat(store.Path(model) + partialSuffix); err != nil {
		t.Error("the partial should be kept to resume from")
	}
}

// A catalogue entry without a usable checksum must fail cleanly, not panic the goroutine.
func TestVerifyRefusesShortChecksums(t *testing.T) {
	path := filepath.Join(t.TempDir(), "m.gguf")
	os.WriteFile(path, []byte("abcd"), 0o644)

	for _, sum := range []string{"", "abc", "not-hex-at-all-but-sixty-four-characters-long-string-of-text!!"} {
		if err := verify(path, sum); err == nil {
			t.Errorf("checksum %q was accepted", sum)
		}
	}
}

func TestLegacyFilesAreListedAndRemoved(t *testing.T) {
	store := testStore(t, nil)
	os.WriteFile(filepath.Join(store.dir, "ggml-base.bin"), []byte("x"), 0o644)
	os.WriteFile(filepath.Join(store.dir, "ggml-small.en.bin"), []byte("x"), 0o644)
	os.WriteFile(filepath.Join(store.dir, "keep.gguf"), []byte("x"), 0o644)

	if got := store.LegacyFiles(); len(got) != 2 {
		t.Fatalf("LegacyFiles = %v, want the two ggml files", got)
	}
	removed, err := store.RemoveLegacyFiles()
	if err != nil || removed != 2 {
		t.Fatalf("RemoveLegacyFiles = %d, %v", removed, err)
	}
	if len(store.LegacyFiles()) != 0 {
		t.Error("legacy files survived")
	}
	if _, err := os.Stat(filepath.Join(store.dir, "keep.gguf")); err != nil {
		t.Error("a gguf model was removed")
	}
}

func TestDeleteRemovesFileAndPartial(t *testing.T) {
	store := testStore(t, nil)
	model := Model{ID: "a/b", Filename: "m.gguf", SizeBytes: 4}

	path := store.Path(model)
	os.MkdirAll(filepath.Dir(path), 0o755)
	os.WriteFile(path, []byte("abcd"), 0o644)
	os.WriteFile(path+partialSuffix, []byte("ab"), 0o644)

	if err := store.Delete(model); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Error("model file survived delete")
	}
	if _, err := os.Stat(path + partialSuffix); !os.IsNotExist(err) {
		t.Error("partial survived delete")
	}
}

func TestDownloadedRequiresFullSize(t *testing.T) {
	store := testStore(t, nil)
	model := Model{ID: "a/b", Filename: "m.gguf", SizeBytes: 100}

	os.WriteFile(store.Path(model), payload(50), 0o644)
	if store.Downloaded(model) {
		t.Error("a short file must not count as downloaded")
	}
	os.WriteFile(store.Path(model), payload(100), 0o644)
	if !store.Downloaded(model) {
		t.Error("a full-size file should count as downloaded")
	}
}
