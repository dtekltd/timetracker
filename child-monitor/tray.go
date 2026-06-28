package main

import (
	_ "embed"

	"child-monitor/internal/logger"
	"child-monitor/internal/services"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// iconData is the tray icon embedded at compile time.
// We use the existing Windows ICO file; systray accepts ICO bytes on Windows.
//
//go:embed build/windows/icon.ico
var iconData []byte

// startTray launches the system tray in a background goroutine.
// It is called from main() after Wails is running.
func (a *App) startTray() {
	go services.RunTray(iconData, services.TrayCallbacks{
		OnShow: func() {
			runtime.Show(a.ctx)
		},
		OnPause: func() {
			if err := a.PauseMonitoring(); err != nil {
				logger.Error("tray pause", err)
			}
			runtime.EventsEmit(a.ctx, "monitoring:paused")
		},
		OnResume: func() {
			if err := a.ResumeMonitoring(); err != nil {
				logger.Error("tray resume", err)
			}
			runtime.EventsEmit(a.ctx, "monitoring:resumed")
		},
		OnOpenSettings: func() {
			runtime.Show(a.ctx)
			runtime.EventsEmit(a.ctx, "nav:settings")
		},
		OnExit: func() {
			// Signal the frontend to show the exit password dialog.
			runtime.Show(a.ctx)
			runtime.EventsEmit(a.ctx, "tray:exit-requested")
		},
		IsPaused: func() bool {
			return a.monitoringPaused
		},
	})
}
