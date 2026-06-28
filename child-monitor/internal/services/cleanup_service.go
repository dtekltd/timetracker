package services

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"child-monitor/internal/db"
	"child-monitor/internal/logger"
)

// CleanupOldScreenshots deletes screenshot files and DB records older than
// the configured retention period. Also removes orphaned DB records where
// the file no longer exists on disk.
func CleanupOldScreenshots() error {
	retentionStr, _ := GetSetting("screenshot_retention_days")
	days, _ := strconv.Atoi(retentionStr)
	if days <= 0 {
		// "Never delete" sentinel.
		logger.Info("screenshot retention: never delete, skipping cleanup")
		return nil
	}

	cutoff := time.Now().AddDate(0, 0, -days).Format(time.DateTime)

	// Fetch records older than cutoff.
	rows, err := db.DB.Query(
		`SELECT id, file_path FROM screenshots WHERE captured_at < ?`, cutoff,
	)
	if err != nil {
		return fmt.Errorf("query old screenshots: %w", err)
	}

	type record struct {
		id   int64
		path string
	}
	var toDelete []record
	for rows.Next() {
		var r record
		if err := rows.Scan(&r.id, &r.path); err != nil {
			rows.Close()
			return err
		}
		toDelete = append(toDelete, r)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return err
	}

	for _, r := range toDelete {
		_ = os.Remove(r.path)
		if _, err := db.DB.Exec(`DELETE FROM screenshots WHERE id = ?`, r.id); err != nil {
			logger.Error(fmt.Sprintf("delete screenshot record id=%d", r.id), err)
		}
	}
	logger.Info(fmt.Sprintf("cleanup: deleted %d screenshot(s) older than %d day(s)", len(toDelete), days))

	// Orphan cleanup: remove records whose files no longer exist.
	if err := cleanupOrphanedRecords(); err != nil {
		logger.Warn("orphan cleanup warning", err)
	}

	return nil
}

// cleanupOrphanedRecords removes database records where the file is missing.
func cleanupOrphanedRecords() error {
	rows, err := db.DB.Query(`SELECT id, file_path FROM screenshots`)
	if err != nil {
		return err
	}

	var orphans []int64
	for rows.Next() {
		var id int64
		var path string
		if err := rows.Scan(&id, &path); err != nil {
			rows.Close()
			return err
		}
		if _, err := os.Stat(path); os.IsNotExist(err) {
			orphans = append(orphans, id)
		}
	}
	rows.Close()

	for _, id := range orphans {
		if _, err := db.DB.Exec(`DELETE FROM screenshots WHERE id = ?`, id); err != nil {
			logger.Error(fmt.Sprintf("delete orphan record id=%d", id), err)
		}
	}
	if len(orphans) > 0 {
		logger.Info(fmt.Sprintf("cleanup: removed %d orphaned screenshot record(s)", len(orphans)))
	}
	return nil
}

// CleanupOldActivityLogs deletes activity samples older than the configured retention.
func CleanupOldActivityLogs() error {
	retentionStr, _ := GetSetting("activity_log_retention_days")
	days, _ := strconv.Atoi(retentionStr)
	if days <= 0 {
		return nil
	}

	cutoff := time.Now().AddDate(0, 0, -days).Format(time.DateTime)
	res, err := db.DB.Exec(`DELETE FROM activity_samples WHERE sampled_at < ?`, cutoff)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n > 0 {
		logger.Info(fmt.Sprintf("cleanup: deleted %d activity sample(s) older than %d day(s)", n, days))
	}
	return nil
}
