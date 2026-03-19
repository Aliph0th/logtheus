package models

import (
	"time"
)

type Application struct {
	ID          uint64  `gorm:"primaryKey" json:"id"`
	Name        string  `gorm:"not null;size:200" json:"name"`
	Description *string `gorm:"not null;size:500" json:"description"`
	ProjectID   uint64  `gorm:"not null" json:"project_id"`

	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}
