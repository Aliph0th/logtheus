package models

import (
	"time"
)

type Project struct {
	ID          uint64  `gorm:"primaryKey" json:"id"`
	Name        string  `gorm:"not null;size:200" json:"name"`
	Description *string `gorm:"not null;size:500" json:"description"`
	OwnerID     uint64  `gorm:"not null" json:"owner_id"`

	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}
