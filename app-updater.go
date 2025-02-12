// Copyright (c) 2024
// Licensed under the MIT License. See LICENSE file in the root directory.

package main

func (a *App) CheckForUpdates() (string, error) {
	return a.updater.CheckForUpdate()
}

func (a *App) PerformUpdate() error {
	return a.updater.PerformUpdate()
}
