package models

type DailySummary struct {
	UsageDate       string `json:"usage_date"`
	TotalSeconds    int    `json:"total_seconds"`
	ActiveSeconds   int    `json:"active_seconds"`
	IdleSeconds     int    `json:"idle_seconds"`
	ScreenshotCount int    `json:"screenshot_count"`
	TopApp          string `json:"top_app"`
}

type DashboardData struct {
	Date              string          `json:"date"`
	Summary           DailySummary    `json:"summary"`
	TopApps           []DailyAppUsage `json:"top_apps"`
	LatestScreenshots []Screenshot    `json:"latest_screenshots"`
}

type AppStatus struct {
	Version             string `json:"version"`
	MonitoringPaused    bool   `json:"monitoring_paused"`
	AutoStartEnabled    bool   `json:"auto_start_enabled"`
	ScreenshotFolder    string `json:"screenshot_folder"`
	ScreenshotServerURL string `json:"screenshot_server_url"`
}
