package models

import (
	"logtheus/application/internal/consts"
	"time"
)

type ApplicationKey struct {
	Signature string              `gorm:"primaryKey;size:64" json:"signature"`
	TokenHash string              `gorm:"not null;size:64" json:"-"`
	Prefix    consts.ApiKeyPrefix `gorm:"not null;size:20" json:"prefix"`

	ApplicationID uint64      `gorm:"not null" json:"application_id"`
	Application   Application `gorm:"foreignKey:ApplicationID;constraint:OnDelete:CASCADE;" json:"-"`

	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
}
