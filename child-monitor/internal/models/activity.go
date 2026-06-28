package models

import "time"

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
