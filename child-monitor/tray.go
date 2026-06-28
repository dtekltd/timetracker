package main

import (
	_ "embed"

	"child-monitor/internal/logger"
	"child-monitor/internal/services"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

//go:embed build/windows/icon.ico
var iconData []byte

// startTray launches the system tray in a background goroutine.
func (a *App) startTray() {
	go services.RunTray(iconData, services.TrayCallbacks{

		// OnShow: show the window — the lock overlay is already set by beforeClose,
		// so the frontend will display the password prompt automatically.
		OnShow: func() {
			runtime.Show(a.ctx)
		},

		// OnPauseRequest: show the window and signal the frontend to ask for a
		// password before pausing. The actual PauseMonitoring() call happens in
		// the frontend after the user confirms the password.
		OnPauseRequest: func() {
			runtime.Show(a.ctx)
			runtime.EventsEmit(a.ctx, "tray:pause-requested")
		},

		// OnResume: resume directly — no password required to resume.
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
			runtime.Show(a.ctx)
			runtime.EventsEmit(a.ctx, "tray:exit-requested")
		},
	})
}
