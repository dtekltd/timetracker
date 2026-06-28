package main

import (
	"embed"
	"net/http"
	"os"
	"strings"

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

	// Set a default password "123123" on first run so the app is usable immediately.
	// Users should change this in Settings.
	if has, _ := services.HasPassword(); !has {
		if err := services.SetPassword("123123"); err != nil {
			logger.Error("set default password", err)
		} else {
			logger.Info("Default password '123123' set — please change it in Settings")
		}
	}

	app := NewApp()

	err := wails.Run(&options.App{
		Title:            "Child Monitor",
		Width:            1200,
		Height:           800,
		MinWidth:         900,
		MinHeight:        600,
		AssetServer: &assetserver.Options{
			Assets:     assets,
			Middleware: screenshotMiddleware,
		},
		BackgroundColour: &options.RGBA{R: 255, G: 255, B: 255, A: 1},
		OnStartup:        app.startup,
		OnDomReady:       app.domReady,
		OnBeforeClose:    app.beforeClose,
		OnShutdown:       app.shutdown,
		Bind:             []interface{}{app},
		Windows: &windows.Options{
			WebviewIsTransparent: false,
			WindowIsTranslucent:  false,
			DisableWindowIcon:    false,
		},
	})
	if err != nil {
		logger.Error("wails run", err)
	}
}

// screenshotMiddleware intercepts GET /api/screenshot?path=<absolute-path> requests
// and serves the file directly from disk. This runs inside the Wails asset server so
// there are no CORS issues — same origin as the frontend.
func screenshotMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/screenshot" {
			next.ServeHTTP(w, r)
			return
		}

		path := r.URL.Query().Get("path")
		if path == "" {
			http.Error(w, "missing path", http.StatusBadRequest)
			return
		}

		// Reject path traversal attempts.
		if strings.Contains(path, "..") {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}

		// File must exist.
		if _, err := os.Stat(path); os.IsNotExist(err) {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}

		http.ServeFile(w, r, path)
	})
}
