package models

import (
	"time"

	"gorm.io/gorm"
)

type Pipeline struct {
	ID        uint           `gorm:"primaryKey;autoIncrement" json:"id"`
	PN        string         `gorm:"type:varchar(255);index" json:"pn"`
	NamaRMFT  string         `gorm:"type:varchar(255);column:nama_rmft" json:"nama_rmft"`
	KodeUker  string         `gorm:"type:varchar(255);index;column:kode_uker" json:"kode_uker"`
	KC        string         `gorm:"type:varchar(255)" json:"kc"`
	Prod      string         `gorm:"type:varchar(255)" json:"prod"`
	NoRek     string         `gorm:"type:varchar(255);column:no_rek" json:"no_rek"`
	Dup       string         `gorm:"type:varchar(255)" json:"dup"`
	Nama      string         `gorm:"type:varchar(255)" json:"nama"`
	TGL       string         `gorm:"type:varchar(255)" json:"tgl"`
	Strategy  string         `gorm:"type:varchar(255);index:idx_strategy" json:"strategy"`
	Segment   string         `gorm:"type:varchar(255);index:idx_segment" json:"segment"`
	Pipeline  string         `gorm:"type:varchar(255);index:idx_pipeline" json:"pipeline"`
	Proyeksi  float64        `gorm:"type:double;default:0;index:idx_proyeksi" json:"proyeksi"`
	CreatedAt time.Time      `gorm:"index:idx_created_at" json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName specifies the table name for Pipeline model
func (Pipeline) TableName() string {
	return "pipelines"
}

// PipelineResponse is used for API responses
type PipelineResponse struct {
	Data       []Pipeline `json:"data"`
	Total      int64      `json:"total"`
	Page       int        `json:"page"`
	PerPage    int        `json:"per_page"`
	TotalPages int        `json:"total_pages"`
}

// ImportResponse is used for CSV import responses
type ImportResponse struct {
	Success    int     `json:"success"`
	Failed     int     `json:"failed"`
	Total      int     `json:"total"`
	Duration   string  `json:"duration"`
	DurationMs int64   `json:"duration_ms"`
	Speed      float64 `json:"speed"`
	Message    string  `json:"message"`
	StartTime  string  `json:"start_time"`
	EndTime    string  `json:"end_time"`
}
