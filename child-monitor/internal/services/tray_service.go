package services

import (
	"child-monitor/internal/logger"

	"github.com/getlantern/systray"
)

// TrayCallbacks holds functions provided by the App layer so the tray can
// trigger UI actions without importing Wails runtime directly.
type TrayCallbacks struct {
	OnShow          func()
	OnPause         func()
	OnResume        func()
	OnOpenSettings  func()
	OnExit          func()
	IsPaused        func() bool
}

// RunTray starts the system tray. This blocks until systray.Quit() is called.
// It must be run from the main goroutine (or a dedicated goroutine on Windows).
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

func onTrayReady(iconData []byte, cb TrayCallbacks) {
	systray.SetIcon(iconData)
	systray.SetTitle("Child Monitor")
	systray.SetTooltip("Child Monitor — Running")

	mOpen := systray.AddMenuItem("Open Dashboard", "Show the main window")
	systray.AddSeparator()
	mPause := systray.AddMenuItem("Pause Monitoring", "Stop capturing screenshots and activity")
	mResume := systray.AddMenuItem("Resume Monitoring", "Resume monitoring")
	mResume.Hide()
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
				if cb.OnPause != nil {
					cb.OnPause()
				}
				mPause.Hide()
				mResume.Show()
				systray.SetTooltip("Child Monitor — Paused")
			case <-mResume.ClickedCh:
				if cb.OnResume != nil {
					cb.OnResume()
				}
				mResume.Hide()
				mPause.Show()
				systray.SetTooltip("Child Monitor — Running")
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

// UpdateTrayPausedState updates the tray tooltip and menu item visibility.
func UpdateTrayPausedState(paused bool) {
	if paused {
		systray.SetTooltip("Child Monitor — Paused")
	} else {
		systray.SetTooltip("Child Monitor — Running")
	}
}
