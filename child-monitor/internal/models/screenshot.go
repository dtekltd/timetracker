package models

import "time"

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
