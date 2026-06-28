package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"time"

	"child-monitor/internal/config"
	"child-monitor/internal/logger"
	"child-monitor/internal/models"
	"child-monitor/internal/services"
	winapi "child-monitor/internal/windows"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

const AppVersion = "1.0.0"

// App is the main application struct exposed to the Wails frontend.
type App struct {
	ctx              context.Context
	cancelWorkers    context.CancelFunc
	monitoringPaused bool
	allowExit        bool // set to true before runtime.Quit so beforeClose allows the quit
}

func NewApp() *App {
	return &App{}
}

// startup is called by Wails when the app is ready.
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

	// Restore monitoring paused state from DB.
	paused, _ := services.GetSetting("monitoring_paused")
	a.monitoringPaused = paused == "true"

	// Sync auto-start registry path in case the exe moved.
	if err := services.SyncAutoStartPath(); err != nil {
		logger.Warn("sync auto-start path", err)
	}

	// Run initial cleanup.
	go func() {
		if err := services.CleanupOldScreenshots(); err != nil {
			logger.Error("startup cleanup", err)
		}
		if err := services.CleanupOldActivityLogs(); err != nil {
			logger.Error("startup activity cleanup", err)
		}
	}()

	// Start background workers.
	workerCtx, cancel := context.WithCancel(context.Background())
	a.cancelWorkers = cancel
	go a.runScreenshotWorker(workerCtx)
	go a.runActivityWorker(workerCtx)
	go a.runSummaryWorker(workerCtx)
	go a.runCleanupWorker(workerCtx)

	// Launch system tray (non-blocking; runs in its own goroutine).
	a.startTray()
}

// domReady is called when the frontend DOM is ready.
func (a *App) domReady(ctx context.Context) {}

// beforeClose is called when the user tries to close the window.
// Returns true (cancel close) unless allowExit is set.
// Emits window:lock-requested so the frontend locks itself before hiding;
// the next time the window is shown from the tray, the lock overlay appears.
func (a *App) beforeClose(ctx context.Context) bool {
	if a.allowExit {
		return false // let Wails proceed with the quit
	}
	runtime.EventsEmit(ctx, "window:lock-requested")
	runtime.Hide(ctx)
	return true // cancel the close — just hide
}

// shutdown is called when Wails is about to quit.
func (a *App) shutdown(ctx context.Context) {
	if a.cancelWorkers != nil {
		a.cancelWorkers()
	}
	services.QuitTray()
	logger.Info("app shutdown")
	logger.Close()
}

// ─── App Status ──────────────────────────────────────────────────────────────

func (a *App) GetAppStatus() (models.AppStatus, error) {
	autoStart, _ := services.IsAutoStartEnabled()
	folder, _ := services.GetSetting("screenshot_folder")
	if folder == "" {
		folder = config.DefaultScreenshotDir()
	}
	return models.AppStatus{
		Version:          AppVersion,
		MonitoringPaused: a.monitoringPaused,
		AutoStartEnabled: autoStart,
		ScreenshotFolder: folder,
	}, nil
}

func (a *App) GetAppVersion() string { return AppVersion }

func (a *App) PauseMonitoring() error {
	a.monitoringPaused = true
	services.UpdateTrayPausedState(true)
	return services.SetSetting("monitoring_paused", "true")
}

func (a *App) ResumeMonitoring() error {
	a.monitoringPaused = false
	services.UpdateTrayPausedState(false)
	return services.SetSetting("monitoring_paused", "false")
}

func (a *App) IsMonitoringPaused() (bool, error) {
	return a.monitoringPaused, nil
}

// RequestExit verifies the password and, if correct, quits the app.
func (a *App) RequestExit(password string) error {
	ok, err := services.VerifyPassword(password)
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("incorrect password")
	}
	// Set flag before Quit so OnBeforeClose allows the quit through.
	a.allowExit = true
	if a.cancelWorkers != nil {
		a.cancelWorkers()
	}
	runtime.Quit(a.ctx)
	return nil
}

func (a *App) HideWindow() error {
	runtime.Hide(a.ctx)
	return nil
}

func (a *App) ShowWindow() error {
	runtime.Show(a.ctx)
	return nil
}

// ─── Password ────────────────────────────────────────────────────────────────

func (a *App) HasPassword() (bool, error)             { return services.HasPassword() }
func (a *App) SetPassword(password string) error      { return services.SetPassword(password) }
func (a *App) VerifyPassword(password string) (bool, error) { return services.VerifyPassword(password) }
func (a *App) ChangePassword(old, newPwd string) error      { return services.ChangePassword(old, newPwd) }

// ─── Settings ────────────────────────────────────────────────────────────────

func (a *App) GetSettings() (map[string]string, error) { return services.GetAllSettings() }
func (a *App) UpdateSetting(key, value string) error   { return services.SetSetting(key, value) }
func (a *App) UpdateSettings(settings map[string]string) error {
	return services.SetSettings(settings)
}

func (a *App) SelectScreenshotFolder() (string, error) {
	folder, err := runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Select Screenshot Folder",
	})
	if err != nil {
		return "", err
	}
	if folder != "" {
		if err := services.SetSetting("screenshot_folder", folder); err != nil {
			return "", err
		}
	}
	return folder, nil
}

func (a *App) OpenScreenshotFolder() error {
	folder, _ := services.GetSetting("screenshot_folder")
	if folder == "" {
		folder = config.DefaultScreenshotDir()
	}
	_ = os.MkdirAll(folder, 0755)
	return exec.Command("explorer", folder).Start()
}

func (a *App) OpenDataFolder() error {
	_ = os.MkdirAll(config.AppDataDir(), 0755)
	return exec.Command("explorer", config.AppDataDir()).Start()
}

// ─── Auto Start ──────────────────────────────────────────────────────────────

func (a *App) EnableAutoStart() error            { return services.EnableAutoStart() }
func (a *App) DisableAutoStart() error           { return services.DisableAutoStart() }
func (a *App) IsAutoStartEnabled() (bool, error) { return services.IsAutoStartEnabled() }

// ─── Screenshots ─────────────────────────────────────────────────────────────

func (a *App) GetScreenshots(date, startTime, endTime string, limit, offset int) ([]models.Screenshot, error) {
	return services.GetScreenshots(date, startTime, endTime, limit, offset)
}

func (a *App) GetScreenshotFileURL(filePath string) (string, error) {
	return "/api/screenshot?path=" + filePath, nil
}

func (a *App) CaptureScreenshotNow() error {
	saved, err := services.CaptureScreenshotNow()
	if err != nil {
		return err
	}
	if len(saved) > 0 {
		runtime.EventsEmit(a.ctx, "screenshot:captured", saved)
	}
	return nil
}

func (a *App) DeleteScreenshot(id int64) error { return services.DeleteScreenshot(id) }
func (a *App) CleanupOldScreenshots() error    { return services.CleanupOldScreenshots() }

// ─── Activity ─────────────────────────────────────────────────────────────────

func (a *App) GetActivityLog(date, search string, limit, offset int) ([]models.ActivitySample, error) {
	return services.GetActivityLog(date, search, limit, offset)
}

func (a *App) GetDailyAppUsage(date string) ([]models.DailyAppUsage, error) {
	return services.GetDailyAppUsage(date)
}

func (a *App) RebuildDailySummary(date string) error { return services.RebuildDailySummary(date) }

// ─── Dashboard ───────────────────────────────────────────────────────────────

func (a *App) GetDashboardData(date string) (models.DashboardData, error) {
	summary, err := services.GetDailySummary(date)
	if err != nil {
		return models.DashboardData{}, err
	}
	topApps, err := services.GetDailyAppUsage(date)
	if err != nil {
		return models.DashboardData{}, err
	}
	if len(topApps) > 10 {
		topApps = topApps[:10]
	}
	latest, err := services.GetLatestScreenshots(date, 6)
	if err != nil {
		return models.DashboardData{}, err
	}
	return models.DashboardData{
		Date:              date,
		Summary:           summary,
		TopApps:           topApps,
		LatestScreenshots: latest,
	}, nil
}

// ─── Background Workers ──────────────────────────────────────────────────────

func (a *App) runScreenshotWorker(ctx context.Context) {
	ticker := time.NewTicker(a.screenshotInterval())
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			ticker.Reset(a.screenshotInterval())
			if a.monitoringPaused {
				continue
			}
			saved, err := services.CaptureScreenshotNow()
			if err != nil {
				logger.Error("screenshot worker", err)
				continue
			}
			if len(saved) > 0 {
				runtime.EventsEmit(a.ctx, "screenshot:captured", saved)
			}
		}
	}
}

func (a *App) runActivityWorker(ctx context.Context) {
	ticker := time.NewTicker(a.activitySampleDuration())
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			ticker.Reset(a.activitySampleDuration())
			if a.monitoringPaused {
				continue
			}
			info := winapi.GetActiveWindowInfo()
			idleSeconds := winapi.GetIdleSeconds()

			threshStr, _ := services.GetSetting("idle_threshold_seconds")
			thresh, _ := strconv.Atoi(threshStr)
			if thresh <= 0 {
				thresh = 180
			}

			if err := services.InsertActivitySample(info, idleSeconds >= thresh, idleSeconds); err != nil {
				logger.Error("activity worker", err)
			}
		}
	}
}

func (a *App) runSummaryWorker(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			today := time.Now().Format("2006-01-02")
			if err := services.RebuildDailySummary(today); err != nil {
				logger.Error("summary worker", err)
			}
			runtime.EventsEmit(a.ctx, "dashboard:updated", today)
		}
	}
}

func (a *App) runCleanupWorker(ctx context.Context) {
	ticker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := services.CleanupOldScreenshots(); err != nil {
				logger.Error("cleanup worker screenshots", err)
			}
			if err := services.CleanupOldActivityLogs(); err != nil {
				logger.Error("cleanup worker activity", err)
			}
		}
	}
}

func (a *App) screenshotInterval() time.Duration {
	s, _ := services.GetSetting("screenshot_interval_minutes")
	mins, _ := strconv.Atoi(s)
	if mins <= 0 {
		mins = 5
	}
	return time.Duration(mins) * time.Minute
}

func (a *App) activitySampleDuration() time.Duration {
	s, _ := services.GetSetting("activity_sample_seconds")
	secs, _ := strconv.Atoi(s)
	if secs <= 0 {
		secs = 10
	}
	return time.Duration(secs) * time.Second
}
