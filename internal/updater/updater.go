package updater

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
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
	"golang.org/x/oauth2"
)

// Updater handles application updates
type Updater struct {
	Owner      string
	Repo       string
	CurrentVer string
	Token      string // GitHub access token for private repos
}

const ASSET_NAME = "myscript"

// NewUpdater creates a new Updater instance
func NewUpdater(owner, repo, currentVer string) *Updater {
	return &Updater{
		Owner:      owner,
		Repo:       repo,
		CurrentVer: currentVer,
	}
}

// SetToken sets the GitHub access token for private repositories
func (u *Updater) SetToken(token string) {
	u.Token = token
}

func (u *Updater) getClient() *github.Client {
	if u.Token == "" {
		return github.NewClient(nil)
	}

	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: u.Token},
	)
	tc := oauth2.NewClient(context.Background(), ts)

	return github.NewClient(tc)
}

// CheckForUpdate checks for available updates
func (u *Updater) CheckForUpdate() (string, error) {
	client := u.getClient()

	release, _, err := client.Repositories.GetLatestRelease(context.Background(), u.Owner, u.Repo)
	if err != nil {
		return "", err
	}

	currentVersion, err := version.NewVersion(u.CurrentVer)
	if err != nil {
		return "", err
	}

	latestVersion, err := version.NewVersion(release.GetTagName())
	if err != nil {
		return "", err
	}

	if latestVersion.GreaterThan(currentVersion) && u.ItHasReleaseAssets(release) {
		return release.GetTagName(), nil
	}

	return "", nil
}

func (u *Updater) ItHasReleaseAssets(release *github.RepositoryRelease) bool {
	var myReleaseAssets []string

	for _, asset := range release.Assets {
		if strings.Contains(asset.GetName(), ASSET_NAME) {
			myReleaseAssets = append(myReleaseAssets, asset.GetName())
		}
	}

	return len(myReleaseAssets) > 0
}

// PerformUpdate executes the update process
func (u *Updater) PerformUpdate() error {
	client := u.getClient()
	release, _, err := client.Repositories.GetLatestRelease(context.Background(), u.Owner, u.Repo)
	if err != nil {
		return err
	}

	assetName := u.getAssetName()
	var asset *github.ReleaseAsset
	for _, a := range release.Assets {
		if a.GetName() == assetName {
			asset = a
			break
		}
	}

	if asset == nil {
		return fmt.Errorf("asset not found: %s", assetName)
	}

	downloadPath := filepath.Join(os.TempDir(), asset.GetName())
	if err := u.downloadFile(client, asset, downloadPath); err != nil {
		return err
	}

	if runtime.GOOS == "windows" {
		return u.handleWindowsUpdate(downloadPath)
	}
	return u.handleUnixUpdate(downloadPath)
}

func (u *Updater) getAssetName() string {
	os := runtime.GOOS
	arch := runtime.GOARCH

	if os == "windows" {
		return fmt.Sprintf(ASSET_NAME+"-windows-%s-installer.exe", arch)
	}

	ext := "tar.gz"
	if os == "darwin" {
		ext = "zip"
	}

	return fmt.Sprintf(ASSET_NAME+"-%s-%s.%s", os, arch, ext)
}

func (u *Updater) downloadFile(client *github.Client, asset *github.ReleaseAsset, downloadPath string) error {
	// Download the asset using authenticated client
	rc, _, err := client.Repositories.DownloadReleaseAsset(
		context.Background(),
		u.Owner,
		u.Repo,
		asset.GetID(),
		client.Client(),
	)
	if err != nil {
		return err
	}
	defer rc.Close()

	out, err := os.Create(downloadPath)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, rc)
	if err != nil {
		return err
	}

	return nil
}

func (u *Updater) handleWindowsUpdate(installerPath string) error {
	cmd := exec.Command(installerPath)
	if err := cmd.Start(); err != nil {
		return err
	}
	os.Exit(0)
	return nil
}

func (u *Updater) handleUnixUpdate(archivePath string) error {
	var err error
	switch {
	case strings.HasSuffix(archivePath, ".zip"):
		err = u.extractZip(archivePath)
	case strings.HasSuffix(archivePath, ".tar.gz"):
		err = u.extractTarGz(archivePath)
	default:
		return fmt.Errorf("unsupported archive format")
	}

	if err != nil {
		return err
	}

	return u.restartApplication()
}

func (u *Updater) extractZip(archivePath string) error {
	r, err := zip.OpenReader(archivePath)
	if err != nil {
		return err
	}
	defer r.Close()

	exePath, err := os.Executable()
	if err != nil {
		return err
	}

	for _, f := range r.File {
		if !f.FileInfo().IsDir() {
			rc, err := f.Open()
			if err != nil {
				return err
			}
			defer rc.Close()

			return u.replaceExecutable(rc, exePath)
		}
	}

	return fmt.Errorf("no files found in archive")
}

func (u *Updater) extractTarGz(archivePath string) error {
	f, err := os.Open(archivePath)
	if err != nil {
		return err
	}
	defer f.Close()

	gzr, err := gzip.NewReader(f)
	if err != nil {
		return err
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)

	exePath, err := os.Executable()
	if err != nil {
		return err
	}

	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		if !hdr.FileInfo().IsDir() {
			return u.replaceExecutable(tr, exePath)
		}
	}

	return fmt.Errorf("no files found in archive")
}

func (u *Updater) replaceExecutable(src io.Reader, exePath string) error {
	tmpPath := exePath + ".tmp"

	// Write new binary
	out, err := os.OpenFile(tmpPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		return err
	}
	defer out.Close()

	if _, err = io.Copy(out, src); err != nil {
		return err
	}

	// Replace existing binary
	if err = os.Rename(tmpPath, exePath); err != nil {
		return err
	}

	return nil
}

func (u *Updater) restartApplication() error {
	exe, err := os.Executable()
	if err != nil {
		return err
	}

	cmd := exec.Command(exe, os.Args[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Start()
	os.Exit(0)

	return nil
}
