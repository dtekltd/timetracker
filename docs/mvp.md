# PROJECT: Child Computer Time Tracker — Windows MVP

You are Claude Code. Build a Windows desktop app MVP for parents to track children's computer usage.

## 1. Tech Stack

Use this stack:

* Backend: Go
* Desktop framework: Wails v2
* Frontend: Vue 3
* UI framework: Quasar
* State management: Pinia
* Local database: SQLite — use `modernc.org/sqlite` (pure Go, no CGo, no MinGW required)
* Screenshot capture: Go package `github.com/kbinani/screenshot`
* System tray: `github.com/getlantern/systray` (standalone, not Wails built-in tray)
* Windows APIs: use `golang.org/x/sys/windows` or syscall where needed
* Target OS for MVP: Windows only

Do not build cloud upload yet. Design the data model so cloud upload can be added later.

### Why these driver choices

* `modernc.org/sqlite` instead of `mattn/go-sqlite3`: pure Go, no CGo dependency, no MinGW/GCC needed on Windows build machine. ~10–20% slower than CGo version but negligible for this app's workload.
* `github.com/getlantern/systray` instead of Wails built-in tray: Wails v2 built-in tray has limited Windows context menu support. `systray` is battle-tested, supports full menu trees, and runs in its own goroutine alongside Wails.

---

# 2. Product Goal

Build a Windows parental monitoring app with these MVP features:

1. Runs in the background.
2. Shows a tray/taskbar icon to reopen the app.
3. Exit app requires parent password.
4. Auto-starts when Windows starts.
5. Takes screenshots every X minutes.
6. Saves screenshots into a default or custom folder.
7. Automatically deletes screenshots older than X days.
8. Tracks total computer usage time per day.
9. Detects the active app/window and logs usage.
10. Shows a daily dashboard.
11. Shows a screenshot gallery sorted by timestamp descending.
12. Allows filtering screenshots by date and time range.
13. Allows viewing app/window activity log by day.
14. Has settings screen with password protection.

---

# 3. App Name

Use app name:

```text
Child Monitor
```

Internal executable name:

```text
child-monitor
```

Default data directory:

```text
%LOCALAPPDATA%\ChildMonitor\
```

Default screenshot directory:

```text
%USERPROFILE%\Pictures\ChildMonitor\Screenshots\
```

Database path:

```text
%LOCALAPPDATA%\ChildMonitor\monitor.db
```

Log file path:

```text
%LOCALAPPDATA%\ChildMonitor\app.log
```

---

# 4. MVP Scope

## Must Have

### Background / Tray

* App should keep running when the window is closed.
* Closing the main window (including Alt+F4) should hide it, not terminate the process.
  * Use Wails `OnBeforeClose` callback to intercept and cancel window close, then call `runtime.Hide`.
  * Test explicitly that Alt+F4 also hides instead of exits.
* Tray icon should have menu:

  * Open Dashboard
  * Pause Monitoring
  * Resume Monitoring
  * Settings
  * Exit
* Exit must require password.
* Settings must require password.
* If password has not been created yet, force the user to create one on first launch.
* Tray tooltip should show current monitoring status: `Child Monitor — Running` or `Child Monitor — Paused`.

### Password

* Store password as a secure hash, never plain text.
* Use bcrypt.
* Create backend methods:

  * `SetPassword(password string) error`
  * `VerifyPassword(password string) (bool, error)`
  * `HasPassword() (bool, error)`
  * `ChangePassword(oldPassword string, newPassword string) error`

### Auto Start

* Add setting:

  * `auto_start_enabled`
* On Windows, implement auto-start using current user registry:

  * `HKCU\Software\Microsoft\Windows\CurrentVersion\Run`
* Store the **absolute path** of the current executable in the registry value.
* On every app startup, check if the registry path matches the current executable path. If not (e.g. exe was moved), update the registry automatically.
* Create backend methods:

  * `EnableAutoStart() error`
  * `DisableAutoStart() error`
  * `IsAutoStartEnabled() (bool, error)`
  * `SyncAutoStartPath() error` — called on startup to fix stale registry path

### DPI Awareness

* Call `SetProcessDpiAwarenessContext(DPI_AWARENESS_CONTEXT_PER_MONITOR_AWARE_V2)` at app startup (before any screenshot).
* This ensures screenshots capture at native resolution on HiDPI/4K monitors and prevents coordinate mismatches.

### Screenshot Capture

* Default interval: 5 minutes.
* User can configure interval from settings:

  * 1 minute
  * 3 minutes
  * 5 minutes
  * 10 minutes
  * 15 minutes
  * 30 minutes
* Save screenshots as JPG.
* JPG quality default: 80. User can configure from settings (50 / 60 / 70 / 80 / 90 / 100).
* Support multiple monitors:

  * For each monitor, save one image.
  * Filename should contain display index.

Filename format:

```text
YYYY-MM-DD_HH-mm-ss_display-0.jpg
YYYY-MM-DD_HH-mm-ss_display-1.jpg
```

Folder structure:

```text
Screenshots/
  2026-06-28/
    2026-06-28_08-00-00_display-0.jpg
    2026-06-28_08-05-00_display-0.jpg
```

After saving screenshot:

* Insert metadata into SQLite table `screenshots`.

#### Screenshot folder resilience

* On each screenshot capture, check if the configured screenshot folder exists. If not, attempt to create it.
* If creation fails, fall back to the default screenshot folder and log a warning.
* If default folder also fails, log the error and skip this capture cycle (do not crash).

### Screenshot Gallery

Screen: `Screenshots`

Features:

* Show grid of screenshots.
* Sort by `captured_at DESC`.
* Filter by:

  * date
  * start time
  * end time
* Click image to open full preview dialog.
* Show timestamp under each image.
* Show display index if multiple monitors.
* Add button: Open Screenshot Folder.

#### Serving images to frontend

* Do **not** load full-size images as base64 for the gallery — this is too memory-intensive.
* Embed a lightweight HTTP file server in Go using `net/http`, listening on `127.0.0.1` with a randomly chosen free port at startup.
* The server serves only the screenshot directory (no other paths).
* Expose the server's base URL to the frontend via a backend method: `GetScreenshotServerURL() string`.
* Frontend constructs image URLs as `<base_url>/<relative_path_from_screenshot_root>`.
* For full preview dialog, the same URL is used (full resolution, lazy-loaded).

### Retention / Auto Delete

Add setting:

```text
screenshot_retention_days
```

Default: 30 days.

Allow options:

* 7 days
* 14 days
* 30 days
* 60 days
* 90 days
* Never delete

Behavior:

* On app startup, run cleanup once.
* Also run cleanup once per day.
* Delete screenshot files older than retention.
* Remove database records for deleted screenshots.
* If a screenshot file is missing on disk but record exists in DB, remove the DB record too (orphan cleanup).
* Do not delete activity logs in MVP unless setting says so.

Optional setting:

```text
activity_log_retention_days
```

Default: 180 days.

### Active Window Tracking

Implement Windows active window tracker.

Every 10 seconds:

* Get foreground window.
* Get window title.
* Get process ID.
* Get process executable name.
* Get full process path if possible.
* Detect idle state using last input time.

Use Windows APIs:

* `GetForegroundWindow`
* `GetWindowTextW`
* `GetWindowThreadProcessId`
* `OpenProcess`
* `QueryFullProcessImageNameW`
* `GetLastInputInfo`

Default:

```text
activity_sample_seconds = 10
idle_threshold_seconds = 180
```

Every sample should be inserted into `activity_samples`.

If user is idle for more than 180 seconds:

* Mark `is_idle = 1`
* Still record last active app/window if available.

### Daily Usage Summary

The app must calculate daily usage:

* Total computer time
* Active time
* Idle time
* Top apps by usage time
* Screenshot count
* App open count
* Last window title per app

Use sample-based calculation:

* Each activity sample represents `activity_sample_seconds`.
* If sample is idle, add to idle time.
* If sample is active, add to active time.
* Group by date and process name.

Create a background aggregator:

* Run every 1 minute.
* Aggregate today's samples into `daily_app_usage`.
* Update `daily_summary`.

Also create manual backend method:

* `RebuildDailySummary(date string) error`

### Dashboard

Screen: `Dashboard`

Features:

* Date picker.
* Cards:

  * Total computer time
  * Active time
  * Idle time
  * Screenshot count
  * Top app
* Chart or visual list:

  * Top apps by total time
* Timeline:

  * Show hourly usage summary if possible.
* Latest screenshots:

  * Show last 6 screenshots of selected date.
* Button:

  * View all screenshots
  * View activity log

### Activity Log

Screen: `Activity Log`

Features:

* Date picker.
* Table grouped by app:

  * App name/process name
  * Total time
  * Active time
  * Idle time
  * Open count
  * Last window title
* Raw timeline section:

  * Time
  * App/process
  * Window title
  * Idle/Active
* Search input:

  * filter by app name or window title

### Settings

Screen: `Settings`

Must require password before access.

Settings:

* Screenshot folder (with folder picker dialog via `wails/v2/pkg/runtime.OpenDirectoryDialog`)
* Screenshot interval
* Screenshot JPG quality
* Screenshot retention days
* Activity sample seconds
* Idle threshold seconds
* Auto start on/off
* Pause monitoring
* Change password
* Open data folder
* Open screenshot folder
* App version (display only, read from build-time constant)

---

# 5. Privacy / Safety Design

This is a parental-control style local monitoring app.

Requirements:

* Store all data locally.
* No network calls in MVP (except localhost HTTP server for serving screenshot images to the frontend UI — this does not leave the machine).
* No cloud upload in MVP.
* No hidden installation.
* No stealth mode.
* The app can run in the background, but it should be visible in the tray.
* Do not implement keylogging.
* Do not record keystrokes.
* Do not record microphone.
* Do not record webcam.
* Do not capture browser passwords or credentials intentionally.

---

# 6. Database Schema

Use SQLite with `modernc.org/sqlite`.

Create migration system in Go. On app startup, run migrations in order. Each migration is versioned.

## settings

```sql
CREATE TABLE IF NOT EXISTS settings (
  key TEXT PRIMARY KEY,
  value TEXT NOT NULL,
  updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
```

## screenshots

```sql
CREATE TABLE IF NOT EXISTS screenshots (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  captured_at DATETIME NOT NULL,
  file_path TEXT NOT NULL,
  file_name TEXT NOT NULL,
  file_size INTEGER DEFAULT 0,
  width INTEGER DEFAULT 0,
  height INTEGER DEFAULT 0,
  display_index INTEGER DEFAULT 0,
  upload_status TEXT DEFAULT 'local_only',
  uploaded_at DATETIME,
  cloud_file_id TEXT,
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_screenshots_captured_at
ON screenshots(captured_at);

CREATE INDEX IF NOT EXISTS idx_screenshots_display_index
ON screenshots(display_index);
```

## activity_samples

```sql
CREATE TABLE IF NOT EXISTS activity_samples (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  sampled_at DATETIME NOT NULL,
  process_name TEXT,
  process_path TEXT,
  window_title TEXT,
  window_handle TEXT,
  is_idle INTEGER DEFAULT 0,
  idle_seconds INTEGER DEFAULT 0,
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_activity_samples_sampled_at
ON activity_samples(sampled_at);

CREATE INDEX IF NOT EXISTS idx_activity_samples_process_name
ON activity_samples(process_name);

CREATE INDEX IF NOT EXISTS idx_activity_samples_sampled_at_process
ON activity_samples(sampled_at, process_name);

CREATE INDEX IF NOT EXISTS idx_activity_samples_is_idle
ON activity_samples(is_idle);
```

## daily_app_usage

```sql
CREATE TABLE IF NOT EXISTS daily_app_usage (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  usage_date DATE NOT NULL,
  process_name TEXT NOT NULL,
  app_name TEXT,
  total_seconds INTEGER DEFAULT 0,
  active_seconds INTEGER DEFAULT 0,
  idle_seconds INTEGER DEFAULT 0,
  open_count INTEGER DEFAULT 0,
  last_window_title TEXT,
  updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
  UNIQUE(usage_date, process_name)
);

CREATE INDEX IF NOT EXISTS idx_daily_app_usage_date
ON daily_app_usage(usage_date);

CREATE INDEX IF NOT EXISTS idx_daily_app_usage_process
ON daily_app_usage(process_name);
```

## daily_summary

```sql
CREATE TABLE IF NOT EXISTS daily_summary (
  usage_date DATE PRIMARY KEY,
  total_seconds INTEGER DEFAULT 0,
  active_seconds INTEGER DEFAULT 0,
  idle_seconds INTEGER DEFAULT 0,
  screenshot_count INTEGER DEFAULT 0,
  top_app TEXT,
  updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
```

### Note on activity_samples volume

At `activity_sample_seconds = 10`: 6 records/min → 360/hour → ~8,640/day.
With `activity_log_retention_days = 180`: up to ~1.55 million records.

The composite index `(sampled_at, process_name)` is required for acceptable query performance on dashboard aggregation. The summary aggregator (runs every 1 minute) pre-aggregates into `daily_app_usage` and `daily_summary` so the dashboard never queries raw samples directly for display.

---

# 7. Backend Go Architecture

Use this folder structure:

```text
.
├── app.go
├── main.go
├── go.mod
├── build/
│   └── windows/
│       └── child-monitor.ico
├── internal/
│   ├── config/
│   │   └── paths.go          # resolve %LOCALAPPDATA%, %USERPROFILE% paths
│   ├── db/
│   │   ├── db.go
│   │   └── migrations.go
│   ├── logger/
│   │   └── logger.go         # file logger with daily rotation
│   ├── models/
│   │   ├── settings.go
│   │   ├── screenshot.go
│   │   ├── activity.go
│   │   └── dashboard.go
│   ├── services/
│   │   ├── settings_service.go
│   │   ├── password_service.go
│   │   ├── screenshot_service.go
│   │   ├── screenshot_server.go  # embedded HTTP file server for images
│   │   ├── activity_tracker.go
│   │   ├── summary_service.go
│   │   ├── cleanup_service.go
│   │   ├── autostart_windows.go
│   │   └── tray_service.go
│   └── windows/
│       ├── active_window_windows.go
│       ├── idle_windows.go
│       └── dpi_windows.go    # SetProcessDpiAwarenessContext call
└── frontend/
    ├── src/
    │   ├── App.vue
    │   ├── main.ts
    │   ├── router/
    │   ├── stores/
    │   ├── pages/
    │   │   ├── DashboardPage.vue
    │   │   ├── ScreenshotsPage.vue
    │   │   ├── ActivityLogPage.vue
    │   │   ├── SettingsPage.vue
    │   │   └── PasswordSetupPage.vue
    │   └── components/
    │       ├── AppLayout.vue
    │       ├── PasswordDialog.vue
    │       ├── ScreenshotGrid.vue
    │       ├── ScreenshotPreviewDialog.vue
    │       ├── UsageCards.vue
    │       └── TopAppsList.vue
```

---

# 8. Backend Models

Create Go structs.

## Screenshot

```go
type Screenshot struct {
    ID           int64     `json:"id"`
    CapturedAt   time.Time `json:"captured_at"`
    FilePath     string    `json:"file_path"`
    FileName     string    `json:"file_name"`
    FileSize     int64     `json:"file_size"`
    Width        int       `json:"width"`
    Height       int       `json:"height"`
    DisplayIndex int       `json:"display_index"`
    UploadStatus string    `json:"upload_status"`
}
```

## ActivitySample

```go
type ActivitySample struct {
    ID           int64     `json:"id"`
    SampledAt    time.Time `json:"sampled_at"`
    ProcessName  string    `json:"process_name"`
    ProcessPath  string    `json:"process_path"`
    WindowTitle  string    `json:"window_title"`
    WindowHandle string    `json:"window_handle"`
    IsIdle       bool      `json:"is_idle"`
    IdleSeconds  int       `json:"idle_seconds"`
}
```

## DailyAppUsage

```go
type DailyAppUsage struct {
    UsageDate       string `json:"usage_date"`
    ProcessName     string `json:"process_name"`
    AppName         string `json:"app_name"`
    TotalSeconds    int    `json:"total_seconds"`
    ActiveSeconds   int    `json:"active_seconds"`
    IdleSeconds     int    `json:"idle_seconds"`
    OpenCount       int    `json:"open_count"`
    LastWindowTitle string `json:"last_window_title"`
}
```

## DailySummary

```go
type DailySummary struct {
    UsageDate       string `json:"usage_date"`
    TotalSeconds    int    `json:"total_seconds"`
    ActiveSeconds   int    `json:"active_seconds"`
    IdleSeconds     int    `json:"idle_seconds"`
    ScreenshotCount int    `json:"screenshot_count"`
    TopApp          string `json:"top_app"`
}
```

## DashboardData

```go
type DashboardData struct {
    Date              string          `json:"date"`
    Summary           DailySummary    `json:"summary"`
    TopApps           []DailyAppUsage `json:"top_apps"`
    LatestScreenshots []Screenshot    `json:"latest_screenshots"`
}
```

## AppStatus

```go
type AppStatus struct {
    Version           string `json:"version"`
    MonitoringPaused  bool   `json:"monitoring_paused"`
    AutoStartEnabled  bool   `json:"auto_start_enabled"`
    ScreenshotFolder  string `json:"screenshot_folder"`
    ScreenshotServerURL string `json:"screenshot_server_url"`
}
```

---

# 9. Backend Methods Exposed to Frontend

Expose these methods through Wails.

## App status

```go
GetAppStatus() (AppStatus, error)
PauseMonitoring() error
ResumeMonitoring() error
IsMonitoringPaused() (bool, error)
RequestExit(password string) error
HideWindow() error
ShowWindow() error
GetScreenshotServerURL() string
GetAppVersion() string
```

## Password

```go
HasPassword() (bool, error)
SetPassword(password string) error
VerifyPassword(password string) (bool, error)
ChangePassword(oldPassword string, newPassword string) error
```

## Settings

```go
GetSettings() (map[string]string, error)
UpdateSetting(key string, value string) error
UpdateSettings(settings map[string]string) error
SelectScreenshotFolder() (string, error)
OpenScreenshotFolder() error
OpenDataFolder() error
```

## Auto Start

```go
EnableAutoStart() error
DisableAutoStart() error
IsAutoStartEnabled() (bool, error)
```

## Screenshots

```go
GetScreenshots(date string, startTime string, endTime string, limit int, offset int) ([]Screenshot, error)
GetScreenshotFileURL(filePath string) (string, error)
CaptureScreenshotNow() error
DeleteScreenshot(id int64) error
CleanupOldScreenshots() error
```

## Activity

```go
GetActivityLog(date string, search string, limit int, offset int) ([]ActivitySample, error)
GetDailyAppUsage(date string) ([]DailyAppUsage, error)
RebuildDailySummary(date string) error
```

## Dashboard

```go
GetDashboardData(date string) (DashboardData, error)
```

---

# 10. Background Workers

Create workers in Go. All workers must:

* Accept a `context.Context` for cancellation.
* Use `select` with `ctx.Done()` to exit cleanly.
* Log all errors but never `panic` or exit the process.
* Continue the loop after a non-fatal error.

## Screenshot worker

* Starts when app starts.
* Runs unless monitoring is paused.
* Reads `screenshot_interval_minutes`.
* Captures screenshots.
* Saves files.
* Inserts metadata.
* Emits frontend event `screenshot:captured` when screenshot is captured.

```go
for {
    select {
    case <-ctx.Done():
        return
    case <-ticker.C:
        if !monitoringPaused {
            if err := CaptureScreenshotNow(); err != nil {
                logger.Error("screenshot capture failed", err)
            }
        }
    }
}
```

## Activity tracker worker

* Starts when app starts.
* Runs unless monitoring is paused.
* Reads `activity_sample_seconds`.
* Every N seconds:

  * Get active window info.
  * Get idle seconds.
  * Insert sample.

```go
for {
    select {
    case <-ctx.Done():
        return
    case <-ticker.C:
        if !monitoringPaused {
            info := windows.GetActiveWindowInfo()
            idleSeconds := windows.GetIdleSeconds()
            isIdle := idleSeconds >= idleThreshold
            if err := InsertActivitySample(info, isIdle, idleSeconds); err != nil {
                logger.Error("activity sample insert failed", err)
            }
        }
    }
}
```

## Summary worker

* Runs every 1 minute.
* Rebuilds today's summary from `activity_samples` into `daily_app_usage` and `daily_summary`.
* Emits frontend event `dashboard:updated`.

## Cleanup worker

* Runs on startup.
* Runs once per day (use a ticker with 24h interval).
* Deletes old screenshots based on `screenshot_retention_days`.
* Removes orphaned DB records (file missing on disk).
* Optionally deletes old activity logs if `activity_log_retention_days` is configured.

---

# 11. Logging

Create a file logger at `%LOCALAPPDATA%\ChildMonitor\app.log`.

Requirements:

* Structured log format: `[YYYY-MM-DD HH:mm:ss] [LEVEL] message`
* Levels: DEBUG, INFO, WARN, ERROR
* Rotate log file daily: keep last 7 log files, delete older ones.
* Log rotation happens on app startup.
* Do not log sensitive data (passwords, password hashes).

Implement in `internal/logger/logger.go`. Expose a package-level logger used by all services.

---

# 12. Frontend UI

Use Quasar layout.

## Navigation

Left menu:

```text
Dashboard
Screenshots
Activity Log
Settings
```

Top bar:

* App title: Child Monitor
* App version (small, e.g. `v1.0.0`)
* Monitoring status badge:

  * Running (green)
  * Paused (yellow)
* Button:

  * Pause / Resume

---

## Dashboard Page

Route:

```text
/
```

Components:

* Date picker
* Usage cards
* Top apps list
* Latest screenshots grid
* Buttons to navigate to screenshots/activity log

Usage cards:

* Total Time
* Active Time
* Idle Time
* Screenshots
* Top App

Format seconds into human-readable form:

```text
4h 25m
58m
10m
```

---

## Screenshots Page

Route:

```text
/screenshots
```

Filters:

* Date
* Start time
* End time

Grid:

* Image thumbnail (loaded via screenshot HTTP server URL, lazy-loaded)
* Timestamp
* Display index

Preview dialog:

* Full image (via screenshot HTTP server URL)
* Timestamp
* File name
* Open file location button
* Delete button

---

## Activity Log Page

Route:

```text
/activity
```

Sections:

1. Top apps table
2. Raw activity timeline

Top apps table columns:

* App
* Total time
* Active time
* Idle time
* Open count
* Last window title

Raw log columns:

* Time
* Process
* Window title
* Status

Filters:

* Date
* Search text

---

## Settings Page

Route:

```text
/settings
```

Before displaying settings, show password dialog.

Settings fields:

* Screenshot folder (with folder picker button)
* Screenshot interval
* Screenshot JPG quality
* Screenshot retention days
* Activity sample seconds
* Idle threshold seconds
* Auto start (toggle)
* Change password
* Open screenshot folder (button)
* Open data folder (button)
* App version (read-only display)

Add buttons:

* Save Settings
* Capture Screenshot Now
* Cleanup Old Screenshots Now

---

## First Launch Flow

If no password exists:

* Redirect to `PasswordSetupPage`.
* User must create password.
* Minimum password length: 6 characters.
* Confirm password required.
* After password created, go to Dashboard.

**Important**: Database migrations must complete before `HasPassword()` is called. Ensure startup sequence is:

1. Initialize logger
2. Run DB migrations
3. Insert default settings if missing
4. Check `HasPassword()`
5. Start background workers
6. Show UI (redirect based on password state)

---

# 13. Default Settings

Insert default settings on first run:

```text
screenshot_interval_minutes = 5
screenshot_retention_days = 30
activity_sample_seconds = 10
idle_threshold_seconds = 180
activity_log_retention_days = 180
auto_start_enabled = true
monitoring_paused = false
jpg_quality = 80
screenshot_folder = <default: %USERPROFILE%\Pictures\ChildMonitor\Screenshots\>
```

---

# 14. Important Implementation Details

## Monitoring paused state

* Persist `monitoring_paused` in the settings table (key: `monitoring_paused`, value: `true`/`false`).
* On app startup, read this value and restore the paused state.
* Workers check this flag before each sample/capture.
* Tray icon tooltip and top bar badge must reflect current state.

## Screenshot file serving

* Do **not** use base64 for gallery images.
* Start a `net/http` file server on `127.0.0.1:<random free port>` at app startup.
* Serve only the screenshot root directory; reject any path traversal attempts (`..`).
* Expose base URL via `GetScreenshotServerURL()` and in `AppStatus`.
* Frontend uses this URL to build `<img src>` directly.

## Screenshot folder resilience

* Before each capture, verify folder exists. Create if missing.
* If custom folder fails, fall back to default folder and log warning.
* If default folder also fails, skip capture and log error.

## Auto-start registry path sync

* When `EnableAutoStart()` is called, write the current executable's absolute path to registry.
* On each startup, call `SyncAutoStartPath()` to compare registry value with `os.Executable()`. Update if different.

## DPI awareness

* Call `SetProcessDpiAwarenessContext(DPI_AWARENESS_CONTEXT_PER_MONITOR_AWARE_V2)` in `main()` before Wails starts.
* This ensures correct screenshot dimensions on HiDPI and multi-monitor setups.

## Timezone

Use local Windows timezone.

Store timestamps in local time for MVP.

## Error Handling

All background workers should:

* Log errors using the file logger.
* Continue running after non-fatal errors.
* Never crash the app due to one failed screenshot or one failed active-window sample.

## Window Close Behavior

* Clicking X (and Alt+F4) must hide the window, not exit the process.
* Use Wails `OnBeforeClose` to intercept and call `runtime.Hide(ctx)`, returning `true` to cancel the close.
* Test both window close button and Alt+F4.

## Exit

When user selects Exit from tray:

* Show password dialog (in tray goroutine or signal frontend to show dialog).
* If password valid:

  * Cancel all background worker contexts.
  * Stop HTTP screenshot server.
  * Quit tray.
  * Call `runtime.Quit(ctx)`.
* If invalid:

  * Show error notification.

---

# 15. Security Notes

* Do not store plain password.
* Use bcrypt hash.
* Store only hash in settings table:

  * key = `password_hash`
* Never expose password hash to frontend.
* Do not implement stealth mode.
* Do not hide process from Task Manager.
* Do not implement keylogging.
* Screenshot HTTP server binds to `127.0.0.1` only, never `0.0.0.0`.
* Screenshot HTTP server must reject path traversal (`../`) attempts.

---

# 16. Future-Ready Design for Cloud Upload

Cloud upload fields are already included in the `screenshots` table schema:

```sql
upload_status TEXT DEFAULT 'local_only',
uploaded_at DATETIME,
cloud_file_id TEXT
```

Possible future statuses:

```text
local_only
pending
uploading
uploaded
failed
```

No upload logic is implemented in MVP. These fields are present for schema-forward compatibility only.

---

# 17. Build Steps

After implementing:

1. Make sure app runs in dev mode.
2. Make sure database is created and migrations run.
3. Make sure first launch password screen appears.
4. Make sure screenshot capture works (test on HiDPI if possible).
5. Make sure active window tracking works.
6. Make sure dashboard displays today's data.
7. Make sure gallery displays screenshots (via HTTP server URL, not base64).
8. Make sure retention cleanup works, including orphan record cleanup.
9. Make sure settings save correctly and persist across restarts.
10. Make sure tray icon works with full context menu.
11. Make sure closing window (X and Alt+F4) hides app.
12. Make sure exit requires password.
13. Make sure monitoring_paused state persists across restart.
14. Make sure auto-start registry path is updated if exe moves.

---

# 18. Acceptance Criteria

The MVP is complete when:

## First Launch

* App opens.
* If no password exists, password setup screen appears.
* User creates password.
* App goes to dashboard.

## Monitoring

* App captures screenshots every configured interval.
* App logs active window every configured sample interval.
* App detects idle state.
* App continues running when window is closed (X or Alt+F4).
* Monitoring paused state persists across app restarts.

## Gallery

* Screenshots appear in descending timestamp order.
* Screenshots load via HTTP server URL (not base64).
* User can filter by date and time.
* User can preview image.

## Dashboard

* User can select a date.
* Dashboard shows:

  * total computer time
  * active time
  * idle time
  * screenshot count
  * top apps
  * latest screenshots

## Activity Log

* User can select a date.
* User can see app usage by app.
* User can see raw activity timeline.
* User can search by app/window title.

## Settings

* Settings require password.
* User can change screenshot folder.
* User can change screenshot interval.
* User can change JPG quality.
* User can change retention days.
* User can enable/disable auto-start.
* User can change password.
* App version is displayed.

## Cleanup

* Screenshots older than retention setting are deleted.
* Database records for deleted screenshots are removed.
* Orphaned DB records (missing files) are also removed.

## Exit Protection

* Exit from tray requires password.
* Invalid password prevents exit.

---

# 19. Development Strategy

Implement in this order:

## Phase 0 — Pre-build Validation (do before writing code)

* Confirm Wails v2 + `systray` can coexist: run a minimal POC with Wails window + systray tray icon + context menu on Windows.
* Confirm `modernc.org/sqlite` builds cleanly with `wails dev` and `wails build`.
* Confirm `runtime.OpenDirectoryDialog` works on Windows.
* Confirm `OnBeforeClose` intercepts both window X button and Alt+F4.

Only proceed to Phase 1 after these POCs pass.

## Phase 1 — Project Setup

* Create Wails project.
* Add Vue 3 + Quasar + Pinia.
* Add `modernc.org/sqlite`.
* Add `github.com/getlantern/systray`.
* Initialize logger (`internal/logger`).
* Add database migration system.
* Add default settings.
* Call `SetProcessDpiAwarenessContext` in main.

## Phase 2 — Password + Settings

* Implement password setup.
* Implement bcrypt.
* Implement settings service.
* Implement settings UI.
* Implement startup sequence (migrations → default settings → HasPassword check).

## Phase 3 — Screenshot Capture

* Implement DPI-aware screenshot service.
* Save screenshots to folder with resilience (folder check, fallback).
* Insert metadata.
* Start embedded HTTP file server.
* Build gallery UI using HTTP server URLs.

## Phase 4 — Activity Tracking

* Implement Windows active-window API.
* Implement idle detection.
* Insert activity samples.
* Build activity log UI.

## Phase 5 — Dashboard

* Implement summary aggregator.
* Implement daily dashboard.
* Implement top apps.
* Implement latest screenshots.

## Phase 6 — Background + Tray

* Implement all background workers with context cancellation.
* Implement pause/resume (persist to DB).
* Implement systray menu.
* Implement close-to-tray (OnBeforeClose).
* Implement exit password.
* Implement auto-start with path sync.

## Phase 7 — Cleanup + Polish

* Implement retention cleanup with orphan record removal.
* Add open folder buttons.
* Add log rotation.
* Add UI loading states.
* Add empty states.
* Show app version in Settings and top bar.
* Test app restart behavior (paused state, settings restore).

---

# 20. Code Quality Requirements

* Keep backend services separated.
* Do not put all logic in `app.go`.
* Use context cancellation for background workers.
* Use file-based structured logging (not just stderr).
* All SQL queries must be parameterized — no string concatenation in queries.
* Avoid global mutable state unless necessary.
* Frontend must use Pinia stores.
* UI must be responsive and clean.
* Use clear error messages.
* Screenshot HTTP server must validate paths — reject any request with `..` in the path.

---

# 21. Final Deliverable

Create a working Windows MVP app with:

* Go backend
* Wails v2 desktop shell
* Vue 3 + Quasar frontend
* SQLite local storage (`modernc.org/sqlite`, no CGo)
* `systray`-based tray icon with full context menu
* DPI-aware background screenshot capture
* Embedded HTTP file server for screenshot image serving
* Active window usage tracking
* Daily dashboard
* Screenshot gallery
* Password-protected settings and exit
* Retention cleanup with orphan record removal
* Structured file logging with rotation

After coding, provide:

1. How to run in dev mode.
2. How to build Windows executable.
3. Short explanation of project structure.
4. Known limitations.
5. Next recommended features.
