package models

import "time"

type Project struct {
	ID          uint    `gorm:"primaryKey;autoIncrement" json:"id"`
	Name        string  `gorm:"not null;size:200" json:"name"`
	Description *string `json:"description"`
	OwnerID     uint    `gorm:"not null" json:"ownerID"`
	Owner       User    `gorm:"foreignKey:OwnerID;constraint:OnDelete:CASCADE;" json:"owner"`

	CreatedAt time.Time `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updatedAt"`
}
