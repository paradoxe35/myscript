// Copyright (c) 2024
// Licensed under the MIT License. See LICENSE file in the root directory.

package updater

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/google/go-github/v50/github"
	"github.com/hashicorp/go-version"
	"github.com/minio/selfupdate"
)

const ASSET_NAME = "myscript"

type Updater struct {
	Owner      string
	Repo       string
	CurrentVer string

	client *github.Client
}

func NewUpdater(owner, repo, currentVer string) *Updater {
	return &Updater{
		Owner:      owner,
		Repo:       repo,
		CurrentVer: currentVer,
		client:     github.NewClient(nil),
	}
}

// Empty when the app is current or the release has no build for this platform.
func (u *Updater) CheckForUpdate() (string, error) {
	release, err := u.latestRelease()
	if err != nil {
		return "", err
	}

	current, err := version.NewVersion(u.CurrentVer)
	if err != nil {
		return "", err
	}

	latest, err := version.NewVersion(release.GetTagName())
	if err != nil {
		return "", err
	}

	if !latest.GreaterThan(current) {
		return "", nil
	}

	// Never announce an update the release cannot deliver.
	if findAsset(release, u.assetName()) == nil {
		return "", nil
	}

	return release.GetTagName(), nil
}

func (u *Updater) PerformUpdate() error {
	release, err := u.latestRelease()
	if err != nil {
		return err
	}

	name := u.assetName()
	asset := findAsset(release, name)
	if asset == nil {
		return fmt.Errorf("this release has no download for %s/%s", runtime.GOOS, runtime.GOARCH)
	}

	if err := u.ensureWritable(); err != nil {
		return err
	}

	// Fetched first so an unverifiable release costs one request, not the transfer.
	checksum, err := u.checksumFor(release, name)
	if err != nil {
		return err
	}

	download := filepath.Join(os.TempDir(), name)
	if err := u.download(asset, download); err != nil {
		return err
	}
	defer os.Remove(download)

	if err := verifyChecksum(download, checksum); err != nil {
		return err
	}

	return u.install(download)
}

func (u *Updater) install(download string) error {
	switch {
	case runtime.GOOS == "windows":
		return runInstaller(download)

	case runtime.GOOS == "darwin":
		if err := installBundle(download); err != nil {
			return err
		}
		return relaunchBundle()

	case runningAsAppImage():
		if err := installAppImage(download); err != nil {
			return err
		}

	default:
		if err := installBinary(download); err != nil {
			return err
		}
	}

	return restart()
}

// The AppImage runtime exports the mounted .AppImage path; the executable
// itself sits on a read-only squashfs, so the outer file is what to replace.
func appImagePath() string { return os.Getenv("APPIMAGE") }

func runningAsAppImage() bool { return appImagePath() != "" }

// The download is already the executable; nothing to unpack.
func installAppImage(download string) error {
	file, err := os.Open(download)
	if err != nil {
		return err
	}
	defer file.Close()

	return selfupdate.Apply(file, selfupdate.Options{
		TargetPath: appImagePath(),
		TargetMode: 0o755,
	})
}

// selfupdate does the atomic rename and rolls back if the swap half-fails.
func installBinary(archivePath string) error {
	binary, closer, err := openBinaryInTarGz(archivePath)
	if err != nil {
		return err
	}
	defer closer.Close()

	return selfupdate.Apply(binary, selfupdate.Options{})
}

// A macOS application is a directory; writing one file over the executable
// would break the bundle and its signature.
func installBundle(archivePath string) error {
	bundle, err := currentBundle()
	if err != nil {
		return err
	}

	staged := bundle + ".new"
	os.RemoveAll(staged)
	defer os.RemoveAll(staged)

	if err := unzipBundle(archivePath, staged); err != nil {
		return err
	}

	previous := bundle + ".old"
	os.RemoveAll(previous)

	if err := os.Rename(bundle, previous); err != nil {
		return err
	}
	if err := os.Rename(staged, bundle); err != nil {
		os.Rename(previous, bundle)
		return err
	}

	os.RemoveAll(previous)
	return nil
}

func currentBundle() (string, error) {
	exe, err := os.Executable()
	if err != nil {
		return "", err
	}
	if exe, err = filepath.EvalSymlinks(exe); err != nil {
		return "", err
	}

	for dir := exe; ; {
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("this build is not inside an .app bundle")
		}
		if strings.HasSuffix(parent, ".app") {
			return parent, nil
		}
		dir = parent
	}
}

// Fails before anything is downloaded when the install is root-owned, as with
// a package manager.
func (u *Updater) ensureWritable() error {
	if runtime.GOOS == "windows" {
		return nil
	}

	target := appImagePath()
	if target == "" {
		var err error
		if target, err = currentBundle(); err != nil {
			if target, err = os.Executable(); err != nil {
				return err
			}
		}
	}

	if err := (&selfupdate.Options{TargetPath: target}).CheckPermissions(); err != nil {
		return fmt.Errorf(
			"%s cannot update itself because %s is not writable — install the new version with your package manager instead",
			ASSET_NAME, target,
		)
	}
	return nil
}

func (u *Updater) assetName() string {
	if runtime.GOOS == "windows" {
		return fmt.Sprintf("%s-windows-%s-installer.exe", ASSET_NAME, runtime.GOARCH)
	}

	if runtime.GOOS == "darwin" {
		return fmt.Sprintf("%s-darwin-%s.zip", ASSET_NAME, runtime.GOARCH)
	}
	if runningAsAppImage() {
		return fmt.Sprintf("%s-linux-%s.AppImage", ASSET_NAME, runtime.GOARCH)
	}

	return fmt.Sprintf("%s-%s-%s.tar.gz", ASSET_NAME, runtime.GOOS, runtime.GOARCH)
}

func (u *Updater) latestRelease() (*github.RepositoryRelease, error) {
	release, _, err := u.client.Repositories.GetLatestRelease(context.Background(), u.Owner, u.Repo)
	return release, err
}

func (u *Updater) checksumFor(release *github.RepositoryRelease, name string) ([]byte, error) {
	asset := findAsset(release, checksumsAsset)
	if asset == nil {
		return nil, fmt.Errorf("this release publishes no %s, so the download cannot be verified", checksumsAsset)
	}

	body, err := u.openAsset(asset)
	if err != nil {
		return nil, err
	}
	defer body.Close()

	sums, err := parseChecksums(body)
	if err != nil {
		return nil, err
	}

	sum, ok := sums[name]
	if !ok {
		return nil, fmt.Errorf("%s does not list %s", checksumsAsset, name)
	}
	return sum, nil
}

func (u *Updater) download(asset *github.ReleaseAsset, path string) error {
	body, err := u.openAsset(asset)
	if err != nil {
		return err
	}
	defer body.Close()

	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = io.Copy(file, body)
	return err
}

func (u *Updater) openAsset(asset *github.ReleaseAsset) (io.ReadCloser, error) {
	body, _, err := u.client.Repositories.DownloadReleaseAsset(
		context.Background(), u.Owner, u.Repo, asset.GetID(), u.client.Client(),
	)
	return body, err
}

func findAsset(release *github.RepositoryRelease, name string) *github.ReleaseAsset {
	for _, asset := range release.Assets {
		if asset.GetName() == name {
			return asset
		}
	}
	return nil
}

func runInstaller(path string) error {
	if err := exec.Command(path).Start(); err != nil {
		return err
	}
	os.Exit(0)
	return nil
}

// Through open, so Launch Services starts it with the bundle's identity
// rather than as a bare child process.
func relaunchBundle() error {
	bundle, err := currentBundle()
	if err != nil {
		return err
	}
	if err := exec.Command("open", "-n", bundle).Start(); err != nil {
		return err
	}
	os.Exit(0)
	return nil
}

func restart() error {
	exe := appImagePath()
	if exe == "" {
		var err error
		if exe, err = os.Executable(); err != nil {
			return err
		}
	}

	command := exec.Command(exe, os.Args[1:]...)
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr
	if err := command.Start(); err != nil {
		return err
	}

	os.Exit(0)
	return nil
}
