// Copyright (c) 2024
// Licensed under the MIT License. See LICENSE file in the root directory.

package updater

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// A package-manager install is root-owned, so the package manager replaces
// it, run through polkit so the user gets the usual password prompt.
type packageManager struct {
	name    string
	suffix  string
	owns    []string
	install []string
}

var packageManagers = []packageManager{
	{"deb", ".deb", []string{"dpkg", "-S"}, []string{"dpkg", "-i"}},
	{"rpm", ".rpm", []string{"rpm", "-qf"}, []string{"rpm", "-U"}},
	{"arch", ".pkg.tar.zst", []string{"pacman", "-Qo"}, []string{"pacman", "-U", "--noconfirm"}},
}

// Swapped in tests, which do not run from a package.
var owningPackageManager = detectPackageManager

func detectPackageManager() *packageManager {
	if runtime.GOOS != "linux" || runningAsAppImage() {
		return nil
	}
	exe, err := os.Executable()
	if err != nil {
		return nil
	}
	if exe, err = filepath.EvalSymlinks(exe); err != nil {
		return nil
	}

	for i := range packageManagers {
		pm := &packageManagers[i]
		if _, err := exec.LookPath(pm.owns[0]); err != nil {
			continue
		}
		if exec.Command(pm.owns[0], append(pm.owns[1:], exe)...).Run() == nil {
			return pm
		}
	}
	return nil
}

func (u *Updater) packaged() *packageManager {
	u.packageOnce.Do(func() { u.packageManager = owningPackageManager() })
	return u.packageManager
}

func (pm *packageManager) assetName() string {
	return fmt.Sprintf("%s-linux-%s%s", ASSET_NAME, runtime.GOARCH, pm.suffix)
}

func (pm *packageManager) installCommand(file string) []string {
	return append(append([]string{"pkexec"}, pm.install...), file)
}

func installPackage(pm *packageManager, file string) error {
	if _, err := exec.LookPath("pkexec"); err != nil {
		return fmt.Errorf("%s was installed as a %s package; install %s with your package manager", ASSET_NAME, pm.name, filepath.Base(file))
	}

	command := pm.installCommand(file)
	out, err := exec.Command(command[0], command[1:]...).CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s failed: %s", strings.Join(command[1:], " "), strings.TrimSpace(string(out)))
	}
	return nil
}
