package main

func (a *App) CheckForUpdates() (string, error) {
	return a.updater.CheckForUpdate()
}

func (a *App) PerformUpdate() error {
	return a.updater.PerformUpdate()
}
