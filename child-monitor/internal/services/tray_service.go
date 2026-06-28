package services

import (
	"child-monitor/internal/logger"

	"github.com/getlantern/systray"
)

// TrayCallbacks holds functions provided by the App layer.
type TrayCallbacks struct {
	OnShow         func()
	OnPauseRequest func() // request pause — frontend handles password
	OnResume       func() // direct resume, no password needed
	OnOpenSettings func()
	OnExit         func()
}

// package-level menu item references so UpdateTrayPausedState can toggle them.
var (
	trayPauseItem  *systray.MenuItem
	trayResumeItem *systray.MenuItem
)

// RunTray starts the system tray. Blocks until QuitTray() is called.
func RunTray(iconData []byte, cb TrayCallbacks) {
	systray.Run(func() {
		onTrayReady(iconData, cb)
	}, func() {
		logger.Info("tray exited")
	})
}

// QuitTray shuts down the system tray.
func QuitTray() {
	systray.Quit()
}

// UpdateTrayPausedState syncs the tray tooltip and menu items with the actual
// monitoring state. Called by PauseMonitoring / ResumeMonitoring in app.go.
func UpdateTrayPausedState(paused bool) {
	if paused {
		systray.SetTooltip("Child Monitor — Paused")
		if trayPauseItem != nil {
			trayPauseItem.Hide()
		}
		if trayResumeItem != nil {
			trayResumeItem.Show()
		}
	} else {
		systray.SetTooltip("Child Monitor — Running")
		if trayResumeItem != nil {
			trayResumeItem.Hide()
		}
		if trayPauseItem != nil {
			trayPauseItem.Show()
		}
	}
}

func onTrayReady(iconData []byte, cb TrayCallbacks) {
	systray.SetIcon(iconData)
	systray.SetTitle("Child Monitor")
	systray.SetTooltip("Child Monitor — Running")

	mOpen := systray.AddMenuItem("Open Dashboard", "Show the main window (password required)")
	systray.AddSeparator()

	mPause := systray.AddMenuItem("Pause Monitoring", "Pause monitoring (password required)")
	mResume := systray.AddMenuItem("Resume Monitoring", "Resume monitoring")
	mResume.Hide()
	trayPauseItem = mPause
	trayResumeItem = mResume

	systray.AddSeparator()
	mSettings := systray.AddMenuItem("Settings", "Open settings (password required)")
	systray.AddSeparator()
	mExit := systray.AddMenuItem("Exit", "Exit the application (password required)")

	go func() {
		for {
			select {
			case <-mOpen.ClickedCh:
				if cb.OnShow != nil {
					cb.OnShow()
				}
			case <-mPause.ClickedCh:
				// Don't pause directly — emit event so frontend asks for password.
				if cb.OnPauseRequest != nil {
					cb.OnPauseRequest()
				}
			case <-mResume.ClickedCh:
				if cb.OnResume != nil {
					cb.OnResume()
				}
			case <-mSettings.ClickedCh:
				if cb.OnOpenSettings != nil {
					cb.OnOpenSettings()
				}
			case <-mExit.ClickedCh:
				if cb.OnExit != nil {
					cb.OnExit()
				}
			}
		}
	}()
}
