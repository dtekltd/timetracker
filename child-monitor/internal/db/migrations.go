package db

import (
	"database/sql"
	"fmt"
)

// RunMigrations creates all tables and indexes if they do not exist.
// This is an additive-only migration; it never drops or alters columns.
func RunMigrations(db *sql.DB) error {
	migrations := []struct {
		name string
		sql  string
	}{
		{"create_settings", sqlCreateSettings},
		{"create_screenshots", sqlCreateScreenshots},
		{"create_activity_samples", sqlCreateActivitySamples},
		{"create_daily_app_usage", sqlCreateDailyAppUsage},
		{"create_daily_summary", sqlCreateDailySummary},
		{"create_schema_version", sqlCreateSchemaVersion},
	}

	for _, m := range migrations {
		if _, err := db.Exec(m.sql); err != nil {
			return fmt.Errorf("migration %q: %w", m.name, err)
		}
	}
	return nil
}

const sqlCreateSettings = `
CREATE TABLE IF NOT EXISTS settings (
  key        TEXT PRIMARY KEY,
  value      TEXT NOT NULL,
  updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);`

const sqlCreateScreenshots = `
CREATE TABLE IF NOT EXISTS screenshots (
  id             INTEGER PRIMARY KEY AUTOINCREMENT,
  captured_at    DATETIME NOT NULL,
  file_path      TEXT NOT NULL,
  file_name      TEXT NOT NULL,
  file_size      INTEGER DEFAULT 0,
  width          INTEGER DEFAULT 0,
  height         INTEGER DEFAULT 0,
  display_index  INTEGER DEFAULT 0,
  upload_status  TEXT DEFAULT 'local_only',
  uploaded_at    DATETIME,
  cloud_file_id  TEXT,
  created_at     DATETIME DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_screenshots_captured_at
  ON screenshots(captured_at);
CREATE INDEX IF NOT EXISTS idx_screenshots_display_index
  ON screenshots(display_index);`

const sqlCreateActivitySamples = `
CREATE TABLE IF NOT EXISTS activity_samples (
  id             INTEGER PRIMARY KEY AUTOINCREMENT,
  sampled_at     DATETIME NOT NULL,
  process_name   TEXT,
  process_path   TEXT,
  window_title   TEXT,
  window_handle  TEXT,
  is_idle        INTEGER DEFAULT 0,
  idle_seconds   INTEGER DEFAULT 0,
  created_at     DATETIME DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_activity_samples_sampled_at
  ON activity_samples(sampled_at);
CREATE INDEX IF NOT EXISTS idx_activity_samples_process_name
  ON activity_samples(process_name);
CREATE INDEX IF NOT EXISTS idx_activity_samples_sampled_at_process
  ON activity_samples(sampled_at, process_name);
CREATE INDEX IF NOT EXISTS idx_activity_samples_is_idle
  ON activity_samples(is_idle);`

const sqlCreateDailyAppUsage = `
CREATE TABLE IF NOT EXISTS daily_app_usage (
  id                INTEGER PRIMARY KEY AUTOINCREMENT,
  usage_date        DATE NOT NULL,
  process_name      TEXT NOT NULL,
  app_name          TEXT,
  total_seconds     INTEGER DEFAULT 0,
  active_seconds    INTEGER DEFAULT 0,
  idle_seconds      INTEGER DEFAULT 0,
  open_count        INTEGER DEFAULT 0,
  last_window_title TEXT,
  updated_at        DATETIME DEFAULT CURRENT_TIMESTAMP,
  UNIQUE(usage_date, process_name)
);
CREATE INDEX IF NOT EXISTS idx_daily_app_usage_date
  ON daily_app_usage(usage_date);
CREATE INDEX IF NOT EXISTS idx_daily_app_usage_process
  ON daily_app_usage(process_name);`

const sqlCreateDailySummary = `
CREATE TABLE IF NOT EXISTS daily_summary (
  usage_date       DATE PRIMARY KEY,
  total_seconds    INTEGER DEFAULT 0,
  active_seconds   INTEGER DEFAULT 0,
  idle_seconds     INTEGER DEFAULT 0,
  screenshot_count INTEGER DEFAULT 0,
  top_app          TEXT,
  updated_at       DATETIME DEFAULT CURRENT_TIMESTAMP
);`

const sqlCreateSchemaVersion = `
CREATE TABLE IF NOT EXISTS schema_version (
  version    INTEGER PRIMARY KEY,
  applied_at DATETIME DEFAULT CURRENT_TIMESTAMP
);`
