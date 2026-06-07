package model

import "time"

type Resource struct {
	ID            uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	Title         string    `gorm:"type:varchar(200);not null;index" json:"title"`
	Summary       string    `gorm:"type:varchar(500)" json:"summary"`
	Category      string    `gorm:"type:varchar(50);index" json:"category"`
	FileName      string    `gorm:"type:varchar(255)" json:"file_name"`
	FilePath      string    `gorm:"type:varchar(500)" json:"file_path"`
	FileSize      int64     `json:"file_size"`
	FileType      string    `gorm:"type:varchar(100)" json:"file_type"`
	FileExt       string    `gorm:"type:varchar(20)" json:"file_ext"`
	CoverURL      string    `gorm:"type:varchar(500)" json:"cover_url"`
	ViewCount     int       `gorm:"type:int;default:0" json:"view_count"`
	DownloadCount int       `gorm:"type:int;default:0" json:"download_count"`
	LikeCount     int       `gorm:"type:int;default:0" json:"like_count"`
	Status        int       `gorm:"type:tinyint;default:1;index" json:"status"`
	IsFeatured    bool      `gorm:"type:boolean;default:false;index" json:"is_featured"`
	CreatedBy     uint      `gorm:"index" json:"created_by"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
	DeletedAt     *time.Time `gorm:"index" json:"-"`
}

func (Resource) TableName() string {
	return "resources"
}
