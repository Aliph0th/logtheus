package models

import (
	"time"
)

type Project struct {
	ID          uint64  `gorm:"primaryKey" json:"id"`
	Name        string  `gorm:"not null;size:200" json:"name"`
	Description *string `json:"description"`
	OwnerID     uint64  `gorm:"not null" json:"ownerID"`

	CreatedAt time.Time `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updatedAt"`
}
