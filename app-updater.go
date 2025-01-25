package main

import "fmt"

func (a *App) CheckForUpdates() (string, error) {
	return a.updater.CheckForUpdate()
}

func (a *App) PerformUpdate() string {
	if err := a.updater.PerformUpdate(); err != nil {
		return fmt.Sprintf("Update failed: %v", err)
	}
	return "Update completed successfully"
}
