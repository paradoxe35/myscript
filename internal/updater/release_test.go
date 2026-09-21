package updater

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/go-github/v50/github"
)

// fakeRelease serves the handful of GitHub endpoints the updater touches.
func fakeRelease(t *testing.T, tag string, assets map[string]string) *Updater {
	t.Helper()

	ids := map[int]string{}
	var listed []string
	id := 1
	for name := range assets {
		ids[id] = name
		listed = append(listed, fmt.Sprintf(
			`{"id":%d,"name":%q,"url":"/repos/o/r/releases/assets/%d"}`, id, name, id,
		))
		id++
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/repos/o/r/releases/latest", func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprintf(w, `{"tag_name":%q,"assets":[%s]}`, tag, strings.Join(listed, ","))
	})
	for assetID, name := range ids {
		body := assets[name]
		mux.HandleFunc(fmt.Sprintf("/repos/o/r/releases/assets/%d", assetID), func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "application/octet-stream")
			w.Write([]byte(body))
		})
	}

	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

	client := github.NewClient(nil)
	client.BaseURL, _ = url.Parse(server.URL + "/")

	return &Updater{Owner: "o", Repo: "r", CurrentVer: "1.0.0", client: client}
}

func checksumsFor(assets map[string]string, skip string) string {
	var lines []string
	for name, body := range assets {
		if name == checksumsAsset || name == skip {
			continue
		}
		sum := sha256.Sum256([]byte(body))
		lines = append(lines, hex.EncodeToString(sum[:])+"  "+name)
	}
	return strings.Join(lines, "\n") + "\n"
}

// Apply() overwrites the running binary without rollback, so a tampered
// download must be rejected before the install step.
func TestPerformUpdateRejectsTamperedAsset(t *testing.T) {
	updater := fakeRelease(t, "v2.0.0", nil)
	name := updater.assetName()

	assets := map[string]string{name: "the genuine release payload"}
	honest := checksumsFor(assets, "")

	assets[name] = "malicious payload swapped in transit"
	assets[checksumsAsset] = honest

	updater = fakeRelease(t, "v2.0.0", assets)

	err := updater.PerformUpdate()
	if err == nil {
		t.Fatal("a tampered asset was installed")
	}
	if !strings.Contains(err.Error(), "corrupt or has been tampered with") {
		t.Fatalf("expected a tamper error, got: %v", err)
	}
}

// A release without checksums.txt must fail closed rather than install blind.
func TestPerformUpdateRequiresChecksums(t *testing.T) {
	updater := fakeRelease(t, "v2.0.0", nil)
	name := updater.assetName()

	updater = fakeRelease(t, "v2.0.0", map[string]string{name: "payload"})

	err := updater.PerformUpdate()
	if err == nil || !strings.Contains(err.Error(), checksumsAsset) {
		t.Fatalf("expected the update to refuse without checksums, got: %v", err)
	}
}

func TestCheckForUpdate(t *testing.T) {
	probe := fakeRelease(t, "v2.0.0", nil)
	name := probe.assetName()

	t.Run("offers a newer release", func(t *testing.T) {
		updater := fakeRelease(t, "v2.0.0", map[string]string{name: "payload"})
		tag, err := updater.CheckForUpdate()
		if err != nil {
			t.Fatal(err)
		}
		if tag != "v2.0.0" {
			t.Fatalf("expected v2.0.0, got %q", tag)
		}
	})

	t.Run("stays quiet when current", func(t *testing.T) {
		updater := fakeRelease(t, "v1.0.0", map[string]string{name: "payload"})
		tag, err := updater.CheckForUpdate()
		if err != nil {
			t.Fatal(err)
		}
		if tag != "" {
			t.Fatalf("offered an update to the version already running: %q", tag)
		}
	})

	// Announcing a release that carries no build for this platform would put
	// the user on a dialog whose only outcome is an error.
	t.Run("stays quiet when this platform has no build", func(t *testing.T) {
		updater := fakeRelease(t, "v2.0.0", map[string]string{"myscript-solaris-sparc.tar.gz": "payload"})
		tag, err := updater.CheckForUpdate()
		if err != nil {
			t.Fatal(err)
		}
		if tag != "" {
			t.Fatalf("offered an update with no asset for this platform: %q", tag)
		}
	})
}

func TestDownloadedArchiveIsCleanedUp(t *testing.T) {
	updater := fakeRelease(t, "v2.0.0", nil)
	name := updater.assetName()
	assets := map[string]string{name: "payload"}
	assets[checksumsAsset] = checksumsFor(assets, "")

	fakeRelease(t, "v2.0.0", assets).PerformUpdate()

	if _, err := os.Stat(filepath.Join(os.TempDir(), name)); !os.IsNotExist(err) {
		t.Fatalf("the downloaded archive was left behind in %s", os.TempDir())
	}
}
