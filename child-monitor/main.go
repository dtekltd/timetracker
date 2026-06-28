package main

import (
	"embed"
	"os"

	"child-monitor/internal/config"
	"child-monitor/internal/db"
	"child-monitor/internal/logger"
	"child-monitor/internal/services"
	winapi "child-monitor/internal/windows"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/windows"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	// Enable per-monitor DPI awareness for correct screenshot dimensions on HiDPI.
	winapi.SetDPIAwareness()

	// Ensure app data directory exists.
	if err := os.MkdirAll(config.AppDataDir(), 0755); err != nil {
		panic("cannot create app data dir: " + err.Error())
	}

	// Initialize file logger.
	if err := logger.Init(config.AppDataDir()); err != nil {
		// Fall through: logger falls back to stderr.
		println("logger init warning:", err.Error())
	}
	logger.Info("Child Monitor starting", "version="+AppVersion)

	// Open and migrate database.
	if err := db.Open(config.DBPath()); err != nil {
		logger.Error("open database", err)
		os.Exit(1)
	}
	defer db.Close()

	if err := db.RunMigrations(db.DB); err != nil {
		logger.Error("run migrations", err)
		os.Exit(1)
	}

	// Insert default settings for any missing keys.
	if err := services.InsertDefaultSettings(); err != nil {
		logger.Error("insert default settings", err)
		os.Exit(1)
	}

	app := NewApp()

	err := wails.Run(&options.App{
		Title:            "Child Monitor",
		Width:            1200,
		Height:           800,
		MinWidth:         900,
		MinHeight:        600,
		AssetServer:      &assetserver.Options{Assets: assets},
		BackgroundColour: &options.RGBA{R: 255, G: 255, B: 255, A: 1},
		OnStartup:        app.startup,
		OnDomReady:       app.domReady,
		OnBeforeClose:    app.beforeClose,
		OnShutdown:       app.shutdown,
		Bind:             []interface{}{app},
		Windows: &windows.Options{
			// Hide window from taskbar when minimised/hidden so tray is the only re-entry.
			WebviewIsTransparent:              false,
			WindowIsTranslucent:               false,
			DisableWindowIcon:                 false,
		},
	})
	if err != nil {
		logger.Error("wails run", err)
	}
}
