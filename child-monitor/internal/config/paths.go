package config

import (
	"os"
	"path/filepath"
)

// AppDataDir returns %LOCALAPPDATA%\ChildMonitor
func AppDataDir() string {
	base := os.Getenv("LOCALAPPDATA")
	if base == "" {
		base = filepath.Join(os.Getenv("USERPROFILE"), "AppData", "Local")
	}
	return filepath.Join(base, "ChildMonitor")
}

// DBPath returns the full path to the SQLite database file.
func DBPath() string {
	return filepath.Join(AppDataDir(), "monitor.db")
}

// LogPath returns the full path to the app log file.
func LogPath() string {
	return filepath.Join(AppDataDir(), "app.log")
}

// DefaultScreenshotDir returns %USERPROFILE%\Pictures\ChildMonitor\Screenshots
func DefaultScreenshotDir() string {
	base := os.Getenv("USERPROFILE")
	return filepath.Join(base, "Pictures", "ChildMonitor", "Screenshots")
}

// EnsureDir creates the directory if it does not exist.
func EnsureDir(path string) error {
	return os.MkdirAll(path, 0755)
}
