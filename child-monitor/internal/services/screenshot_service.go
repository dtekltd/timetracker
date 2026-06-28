package services

import (
	"fmt"
	"image/jpeg"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"child-monitor/internal/config"
	"child-monitor/internal/db"
	"child-monitor/internal/logger"
	"child-monitor/internal/models"

	"github.com/kbinani/screenshot"
)

// CaptureScreenshotNow captures all monitors and saves them as JPG files.
// Returns the list of saved screenshots.
func CaptureScreenshotNow() ([]models.Screenshot, error) {
	folder, err := resolveScreenshotFolder()
	if err != nil {
		return nil, fmt.Errorf("resolve screenshot folder: %w", err)
	}

	qualityStr, _ := GetSetting("jpg_quality")
	quality, _ := strconv.Atoi(qualityStr)
	if quality <= 0 || quality > 100 {
		quality = 80
	}

	now := time.Now()
	dateDir := filepath.Join(folder, now.Format("2006-01-02"))
	if err := os.MkdirAll(dateDir, 0755); err != nil {
		return nil, fmt.Errorf("create date dir: %w", err)
	}

	numDisplays := screenshot.NumActiveDisplays()
	var saved []models.Screenshot

	for i := 0; i < numDisplays; i++ {
		bounds := screenshot.GetDisplayBounds(i)
		img, err := screenshot.CaptureRect(bounds)
		if err != nil {
			logger.Error(fmt.Sprintf("capture display %d failed", i), err)
			continue
		}

		fileName := fmt.Sprintf("%s_display-%d.jpg", now.Format("2006-01-02_15-04-05"), i)
		filePath := filepath.Join(dateDir, fileName)

		f, err := os.Create(filePath)
		if err != nil {
			logger.Error(fmt.Sprintf("create screenshot file display %d", i), err)
			continue
		}

		if err := jpeg.Encode(f, img, &jpeg.Options{Quality: quality}); err != nil {
			f.Close()
			logger.Error("jpeg encode failed", err)
			continue
		}
		f.Close()

		info, err := os.Stat(filePath)
		if err != nil {
			logger.Warn("stat screenshot file failed", err)
		}

		var fileSize int64
		if info != nil {
			fileSize = info.Size()
		}

		s := models.Screenshot{
			CapturedAt:   now,
			FilePath:     filePath,
			FileName:     fileName,
			FileSize:     fileSize,
			Width:        bounds.Dx(),
			Height:       bounds.Dy(),
			DisplayIndex: i,
			UploadStatus: "local_only",
		}

		id, err := insertScreenshot(s)
		if err != nil {
			logger.Error("insert screenshot record", err)
		} else {
			s.ID = id
		}
		saved = append(saved, s)
	}
	return saved, nil
}

// resolveScreenshotFolder returns the configured or default screenshot folder,
// creating it if necessary. Falls back to default if custom folder fails.
func resolveScreenshotFolder() (string, error) {
	folder, _ := GetSetting("screenshot_folder")
	if folder == "" {
		folder = config.DefaultScreenshotDir()
	}

	if err := os.MkdirAll(folder, 0755); err != nil {
		logger.Warn(fmt.Sprintf("custom screenshot folder %q failed, falling back to default", folder))
		folder = config.DefaultScreenshotDir()
		if err2 := os.MkdirAll(folder, 0755); err2 != nil {
			return "", fmt.Errorf("default screenshot folder: %w", err2)
		}
	}
	return folder, nil
}

func insertScreenshot(s models.Screenshot) (int64, error) {
	res, err := db.DB.Exec(
		`INSERT INTO screenshots
		 (captured_at, file_path, file_name, file_size, width, height, display_index, upload_status)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		s.CapturedAt.Format(time.DateTime),
		s.FilePath, s.FileName, s.FileSize, s.Width, s.Height, s.DisplayIndex, s.UploadStatus,
	)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

// GetScreenshots returns screenshots filtered by date and optional time range.
func GetScreenshots(date, startTime, endTime string, limit, offset int) ([]models.Screenshot, error) {
	if limit <= 0 {
		limit = 50
	}

	// Build date range from date + optional start/end time.
	from := date + " 00:00:00"
	to := date + " 23:59:59"
	if startTime != "" {
		from = date + " " + startTime
	}
	if endTime != "" {
		to = date + " " + endTime
	}

	rows, err := db.DB.Query(
		`SELECT id, captured_at, file_path, file_name, file_size, width, height, display_index, upload_status
		 FROM screenshots
		 WHERE captured_at >= ? AND captured_at <= ?
		 ORDER BY captured_at DESC
		 LIMIT ? OFFSET ?`,
		from, to, limit, offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanScreenshots(rows)
}

// GetLatestScreenshots returns the most recent N screenshots for a given date.
func GetLatestScreenshots(date string, n int) ([]models.Screenshot, error) {
	rows, err := db.DB.Query(
		`SELECT id, captured_at, file_path, file_name, file_size, width, height, display_index, upload_status
		 FROM screenshots
		 WHERE captured_at >= ? AND captured_at <= ?
		 ORDER BY captured_at DESC
		 LIMIT ?`,
		date+" 00:00:00", date+" 23:59:59", n,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanScreenshots(rows)
}

func scanScreenshots(rows interface {
	Next() bool
	Scan(...any) error
	Err() error
}) ([]models.Screenshot, error) {
	var list []models.Screenshot
	for rows.Next() {
		var s models.Screenshot
		var capturedAt string
		if err := rows.Scan(&s.ID, &capturedAt, &s.FilePath, &s.FileName,
			&s.FileSize, &s.Width, &s.Height, &s.DisplayIndex, &s.UploadStatus); err != nil {
			return nil, err
		}
		s.CapturedAt, _ = time.ParseInLocation(time.DateTime, capturedAt, time.Local)
		list = append(list, s)
	}
	return list, rows.Err()
}

// DeleteScreenshot removes the file and database record for a screenshot.
func DeleteScreenshot(id int64) error {
	var filePath string
	err := db.DB.QueryRow(`SELECT file_path FROM screenshots WHERE id = ?`, id).Scan(&filePath)
	if err != nil {
		return err
	}
	_ = os.Remove(filePath) // best-effort file deletion
	_, err = db.DB.Exec(`DELETE FROM screenshots WHERE id = ?`, id)
	return err
}

// GetScreenshotFileURL converts an absolute file path to a URL served by
// the local screenshot HTTP server.
func GetScreenshotFileURL(serverURL, filePath string) string {
	// The screenshot server serves the root screenshot dir.
	// The frontend will construct these URLs via GetScreenshotServerURL.
	return serverURL + "/file?path=" + filePath
}
