package services

import (
	"database/sql"
	"fmt"
	"time"

	"child-monitor/internal/db"
)

// Default setting values inserted on first run.
var defaultSettings = map[string]string{
	"screenshot_interval_minutes":  "5",
	"screenshot_retention_days":    "30",
	"activity_sample_seconds":      "10",
	"idle_threshold_seconds":       "180",
	"activity_log_retention_days":  "180",
	"auto_start_enabled":           "true",
	"monitoring_paused":            "false",
	"jpg_quality":                  "80",
	"screenshot_folder":            "", // empty = use default path at runtime
}

// InsertDefaultSettings writes any missing keys with their default values.
func InsertDefaultSettings() error {
	for key, value := range defaultSettings {
		_, err := db.DB.Exec(
			`INSERT OR IGNORE INTO settings (key, value, updated_at) VALUES (?, ?, ?)`,
			key, value, time.Now().Format(time.DateTime),
		)
		if err != nil {
			return fmt.Errorf("insert default setting %q: %w", key, err)
		}
	}
	return nil
}

// GetSetting returns one setting value by key.
func GetSetting(key string) (string, error) {
	var value string
	err := db.DB.QueryRow(`SELECT value FROM settings WHERE key = ?`, key).Scan(&value)
	if err == sql.ErrNoRows {
		return "", nil
	}
	return value, err
}

// GetAllSettings returns all settings as a map.
func GetAllSettings() (map[string]string, error) {
	rows, err := db.DB.Query(`SELECT key, value FROM settings`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[string]string)
	for rows.Next() {
		var k, v string
		if err := rows.Scan(&k, &v); err != nil {
			return nil, err
		}
		// Never expose password hash to frontend.
		if k == "password_hash" {
			continue
		}
		result[k] = v
	}
	return result, rows.Err()
}

// SetSetting upserts a single setting.
func SetSetting(key, value string) error {
	_, err := db.DB.Exec(
		`INSERT INTO settings (key, value, updated_at) VALUES (?, ?, ?)
		 ON CONFLICT(key) DO UPDATE SET value = excluded.value, updated_at = excluded.updated_at`,
		key, value, time.Now().Format(time.DateTime),
	)
	return err
}

// SetSettings upserts multiple settings in one transaction.
func SetSettings(settings map[string]string) error {
	tx, err := db.DB.Begin()
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	stmt, err := tx.Prepare(
		`INSERT INTO settings (key, value, updated_at) VALUES (?, ?, ?)
		 ON CONFLICT(key) DO UPDATE SET value = excluded.value, updated_at = excluded.updated_at`,
	)
	if err != nil {
		return err
	}
	defer stmt.Close()

	now := time.Now().Format(time.DateTime)
	for k, v := range settings {
		if k == "password_hash" {
			continue // safety: never allow frontend to overwrite hash directly
		}
		if _, err := stmt.Exec(k, v, now); err != nil {
			return err
		}
	}
	return tx.Commit()
}
