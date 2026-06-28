package services

import (
	"fmt"
	"strconv"
	"time"

	"child-monitor/internal/db"
	"child-monitor/internal/logger"
	"child-monitor/internal/models"
	winapi "child-monitor/internal/windows"
)

// InsertActivitySample records one activity sample to the database.
func InsertActivitySample(info winapi.ActiveWindowInfo, isIdle bool, idleSeconds int) error {
	isIdleInt := 0
	if isIdle {
		isIdleInt = 1
	}
	_, err := db.DB.Exec(
		`INSERT INTO activity_samples
		 (sampled_at, process_name, process_path, window_title, window_handle, is_idle, idle_seconds)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		time.Now().Format(time.DateTime),
		info.ProcessName, info.ProcessPath, info.WindowTitle, info.WindowHandle,
		isIdleInt, idleSeconds,
	)
	return err
}

// GetActivityLog returns raw activity samples for a date, with optional search.
func GetActivityLog(date, search string, limit, offset int) ([]models.ActivitySample, error) {
	if limit <= 0 {
		limit = 200
	}

	query := `SELECT id, sampled_at, process_name, process_path, window_title, window_handle, is_idle, idle_seconds
	          FROM activity_samples
	          WHERE sampled_at >= ? AND sampled_at <= ?`
	args := []any{date + " 00:00:00", date + " 23:59:59"}

	if search != "" {
		query += ` AND (process_name LIKE ? OR window_title LIKE ?)`
		pattern := "%" + search + "%"
		args = append(args, pattern, pattern)
	}

	query += ` ORDER BY sampled_at ASC LIMIT ? OFFSET ?`
	args = append(args, limit, offset)

	rows, err := db.DB.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []models.ActivitySample
	for rows.Next() {
		var s models.ActivitySample
		var sampledAt string
		var isIdleInt int
		if err := rows.Scan(&s.ID, &sampledAt, &s.ProcessName, &s.ProcessPath,
			&s.WindowTitle, &s.WindowHandle, &isIdleInt, &s.IdleSeconds); err != nil {
			return nil, err
		}
		s.SampledAt, _ = time.ParseInLocation(time.DateTime, sampledAt, time.Local)
		s.IsIdle = isIdleInt == 1
		list = append(list, s)
	}
	return list, rows.Err()
}

// GetDailyAppUsage returns aggregated usage per app for a date.
func GetDailyAppUsage(date string) ([]models.DailyAppUsage, error) {
	rows, err := db.DB.Query(
		`SELECT usage_date, process_name, COALESCE(app_name,''), total_seconds, active_seconds, idle_seconds,
		        open_count, COALESCE(last_window_title,'')
		 FROM daily_app_usage
		 WHERE usage_date = ?
		 ORDER BY total_seconds DESC`,
		date,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []models.DailyAppUsage
	for rows.Next() {
		var u models.DailyAppUsage
		if err := rows.Scan(&u.UsageDate, &u.ProcessName, &u.AppName,
			&u.TotalSeconds, &u.ActiveSeconds, &u.IdleSeconds,
			&u.OpenCount, &u.LastWindowTitle); err != nil {
			return nil, err
		}
		list = append(list, u)
	}
	return list, rows.Err()
}

// RebuildDailySummary aggregates activity_samples for the given date into
// daily_app_usage and daily_summary tables.
func RebuildDailySummary(date string) error {
	sampleSecondsStr, _ := GetSetting("activity_sample_seconds")
	sampleSeconds, _ := strconv.Atoi(sampleSecondsStr)
	if sampleSeconds <= 0 {
		sampleSeconds = 10
	}

	// Aggregate per-app totals from raw samples.
	rows, err := db.DB.Query(
		`SELECT process_name,
		        COUNT(*) AS samples,
		        SUM(CASE WHEN is_idle = 0 THEN 1 ELSE 0 END) AS active_samples,
		        SUM(CASE WHEN is_idle = 1 THEN 1 ELSE 0 END) AS idle_samples,
		        MAX(window_title) AS last_title
		 FROM activity_samples
		 WHERE sampled_at >= ? AND sampled_at <= ?
		 GROUP BY process_name`,
		date+" 00:00:00", date+" 23:59:59",
	)
	if err != nil {
		return fmt.Errorf("aggregate samples: %w", err)
	}
	defer rows.Close()

	type appRow struct {
		processName string
		total, active, idle int
		lastTitle   string
	}

	var apps []appRow
	totalSamples, activeSamples, idleSamples := 0, 0, 0

	for rows.Next() {
		var r appRow
		var samples, active, idle int
		if err := rows.Scan(&r.processName, &samples, &active, &idle, &r.lastTitle); err != nil {
			return err
		}
		r.total = samples * sampleSeconds
		r.active = active * sampleSeconds
		r.idle = idle * sampleSeconds
		apps = append(apps, r)
		totalSamples += samples
		activeSamples += active
		idleSamples += idle
	}
	if err := rows.Err(); err != nil {
		return err
	}

	// Upsert daily_app_usage rows.
	for _, a := range apps {
		_, err := db.DB.Exec(
			`INSERT INTO daily_app_usage
			 (usage_date, process_name, total_seconds, active_seconds, idle_seconds, last_window_title, updated_at)
			 VALUES (?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
			 ON CONFLICT(usage_date, process_name) DO UPDATE SET
			   total_seconds     = excluded.total_seconds,
			   active_seconds    = excluded.active_seconds,
			   idle_seconds      = excluded.idle_seconds,
			   last_window_title = excluded.last_window_title,
			   updated_at        = CURRENT_TIMESTAMP`,
			date, a.processName, a.total, a.active, a.idle, a.lastTitle,
		)
		if err != nil {
			logger.Error("upsert daily_app_usage", err)
		}
	}

	// Determine top app.
	var topApp string
	if len(apps) > 0 {
		topApp = apps[0].processName
		for _, a := range apps {
			if a.active > apps[0].active {
				topApp = a.processName
			}
		}
	}

	// Count screenshots for the day.
	var screenshotCount int
	_ = db.DB.QueryRow(
		`SELECT COUNT(*) FROM screenshots WHERE captured_at >= ? AND captured_at <= ?`,
		date+" 00:00:00", date+" 23:59:59",
	).Scan(&screenshotCount)

	// Upsert daily_summary.
	_, err = db.DB.Exec(
		`INSERT INTO daily_summary
		 (usage_date, total_seconds, active_seconds, idle_seconds, screenshot_count, top_app, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
		 ON CONFLICT(usage_date) DO UPDATE SET
		   total_seconds    = excluded.total_seconds,
		   active_seconds   = excluded.active_seconds,
		   idle_seconds     = excluded.idle_seconds,
		   screenshot_count = excluded.screenshot_count,
		   top_app          = excluded.top_app,
		   updated_at       = CURRENT_TIMESTAMP`,
		date,
		totalSamples*sampleSeconds,
		activeSamples*sampleSeconds,
		idleSamples*sampleSeconds,
		screenshotCount,
		topApp,
	)
	return err
}

// GetDailySummary returns the aggregated summary for a date.
func GetDailySummary(date string) (models.DailySummary, error) {
	var s models.DailySummary
	err := db.DB.QueryRow(
		`SELECT usage_date, total_seconds, active_seconds, idle_seconds, screenshot_count, COALESCE(top_app,'')
		 FROM daily_summary WHERE usage_date = ?`, date,
	).Scan(&s.UsageDate, &s.TotalSeconds, &s.ActiveSeconds, &s.IdleSeconds, &s.ScreenshotCount, &s.TopApp)
	if err != nil {
		// Return empty summary rather than error when no data yet.
		s.UsageDate = date
		return s, nil
	}
	return s, nil
}
