package updater

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestAssetNameFollowsInstallKind(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("the AppImage runtime only exists on Linux")
	}

	updater := &Updater{}

	t.Setenv("APPIMAGE", "")
	if got := updater.assetName(); got != "myscript-linux-"+runtime.GOARCH+".tar.gz" {
		t.Fatalf("a plain install should fetch the tarball, got %q", got)
	}

	t.Setenv("APPIMAGE", "/home/user/Apps/MyScript.AppImage")
	if got := updater.assetName(); got != "myscript-linux-"+runtime.GOARCH+".AppImage" {
		t.Fatalf("an AppImage install should fetch the AppImage, got %q", got)
	}
}

func TestPackagedInstallsFetchTheirPackage(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("packages only exist on Linux")
	}
	t.Setenv("APPIMAGE", "")

	for _, pm := range packageManagers {
		owningPackageManager = func() *packageManager { return &pm }
		t.Cleanup(func() { owningPackageManager = detectPackageManager })

		got := (&Updater{}).assetName()
		if want := "myscript-linux-" + runtime.GOARCH + pm.suffix; got != want {
			t.Errorf("%s install should fetch %q, got %q", pm.name, want, got)
		}
	}
}

func TestPackagesAreInstalledThroughPolkit(t *testing.T) {
	got := packageManagers[0].installCommand("/tmp/myscript-linux-amd64.deb")
	want := "pkexec dpkg -i /tmp/myscript-linux-amd64.deb"
	if strings.Join(got, " ") != want {
		t.Fatalf("got %q, want %q", strings.Join(got, " "), want)
	}
}

// An AppImage runs from a read-only squashfs, so replacing the executable
// reported by os.Executable() would fail. The outer file is the real target.
func TestInstallAppImageReplacesTheOuterFile(t *testing.T) {
	dir := t.TempDir()

	appImage := filepath.Join(dir, "MyScript.AppImage")
	if err := os.WriteFile(appImage, []byte("old release"), 0o755); err != nil {
		t.Fatal(err)
	}

	download := filepath.Join(dir, "download")
	if err := os.WriteFile(download, []byte("new release"), 0o644); err != nil {
		t.Fatal(err)
	}

	t.Setenv("APPIMAGE", appImage)
	if err := installAppImage(download); err != nil {
		t.Fatal(err)
	}

	body, err := os.ReadFile(appImage)
	if err != nil {
		t.Fatal(err)
	}
	if string(body) != "new release" {
		t.Fatalf("the AppImage was not replaced, it holds %q", body)
	}

	// A downloaded asset arrives without the executable bit; losing it would
	// leave the user with an app that no longer starts.
	info, err := os.Stat(appImage)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm()&0o111 == 0 {
		t.Fatalf("the replaced AppImage is not executable: %v", info.Mode())
	}
}

func TestEnsureWritableTargetsTheAppImage(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("the AppImage runtime only exists on Linux")
	}

	readOnly := filepath.Join(t.TempDir(), "locked")
	if err := os.Mkdir(readOnly, 0o555); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Chmod(readOnly, 0o755) })

	t.Setenv("APPIMAGE", filepath.Join(readOnly, "MyScript.AppImage"))

	err := (&Updater{}).ensureWritable()
	if err == nil {
		t.Fatal("an unwritable AppImage location was accepted")
	}
	if !strings.Contains(err.Error(), "not writable") {
		t.Fatalf("expected actionable advice, got: %v", err)
	}
}
