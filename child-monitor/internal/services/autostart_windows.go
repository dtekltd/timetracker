package services

import (
	"fmt"
	"os"

	"golang.org/x/sys/windows/registry"
)

const (
	autoStartKey  = `Software\Microsoft\Windows\CurrentVersion\Run`
	autoStartName = "ChildMonitor"
)

// EnableAutoStart writes the current executable path to the Run registry key.
func EnableAutoStart() error {
	exePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("get executable path: %w", err)
	}
	k, err := registry.OpenKey(registry.CURRENT_USER, autoStartKey, registry.SET_VALUE)
	if err != nil {
		return fmt.Errorf("open registry key: %w", err)
	}
	defer k.Close()
	return k.SetStringValue(autoStartName, exePath)
}

// DisableAutoStart removes the app from the Run registry key.
func DisableAutoStart() error {
	k, err := registry.OpenKey(registry.CURRENT_USER, autoStartKey, registry.SET_VALUE)
	if err != nil {
		return fmt.Errorf("open registry key: %w", err)
	}
	defer k.Close()
	err = k.DeleteValue(autoStartName)
	if err == registry.ErrNotExist {
		return nil
	}
	return err
}

// IsAutoStartEnabled returns true if the registry Run value exists.
func IsAutoStartEnabled() (bool, error) {
	k, err := registry.OpenKey(registry.CURRENT_USER, autoStartKey, registry.QUERY_VALUE)
	if err != nil {
		return false, fmt.Errorf("open registry key: %w", err)
	}
	defer k.Close()
	_, _, err = k.GetStringValue(autoStartName)
	if err == registry.ErrNotExist {
		return false, nil
	}
	return err == nil, err
}

// SyncAutoStartPath updates the registry path if the exe has moved.
// Called on every startup when auto-start is enabled.
func SyncAutoStartPath() error {
	enabled, err := IsAutoStartEnabled()
	if err != nil || !enabled {
		return err
	}

	exePath, err := os.Executable()
	if err != nil {
		return err
	}

	k, err := registry.OpenKey(registry.CURRENT_USER, autoStartKey, registry.QUERY_VALUE|registry.SET_VALUE)
	if err != nil {
		return err
	}
	defer k.Close()

	current, _, _ := k.GetStringValue(autoStartName)
	if current != exePath {
		return k.SetStringValue(autoStartName, exePath)
	}
	return nil
}
