package dto

import "time"

type ResourceListItem struct {
	ID            uint   `json:"id"`
	Title         string `json:"title"`
	Summary       string `json:"summary"`
	Category      string `json:"category"`
	CoverURL      string `json:"cover_url"`
	DownloadCount int    `json:"download_count"`
	ViewCount     int    `json:"view_count"`
}

type ResourceDetail struct {
	ID            uint   `json:"id"`
	Title         string `json:"title"`
	Summary       string `json:"summary"`
	Content       string `json:"content"`
	Category      string `json:"category"`
	FileName      string `json:"file_name"`
	FileSize      int64  `json:"file_size"`
	FileType      string `json:"file_type"`
	CoverURL      string `json:"cover_url"`
	DownloadCount int    `json:"download_count"`
	ViewCount     int    `json:"view_count"`
	CreatedAt     string `json:"created_at"`
}

type MyDownloadItem struct {
	ID            uint      `json:"id"`
	ResourceID    uint      `json:"resource_id"`
	ResourceTitle string    `json:"resource_title"`
	FileName      string    `json:"file_name"`
	FileSize      int64     `json:"file_size"`
	CoverURL      string    `json:"cover_url"`
	IP            string    `json:"ip"`
	CreatedAt     time.Time `json:"created_at"`
}

