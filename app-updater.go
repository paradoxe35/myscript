// Copyright (c) 2024
// Licensed under the MIT License. See LICENSE file in the root directory.

package main

// CheckForUpdates returns the newer tag, or "" when this build is current.
func (a *App) CheckForUpdates() (string, error) {
	release, found, err := a.updater.Check(a.ctx)
	if err != nil || !found {
		return "", err
	}
	return release.Tag, nil
}

// PerformUpdate relaunches into the new version and only returns on failure.
func (a *App) PerformUpdate() error {
	return a.updater.Update(a.ctx, nil)
}
