package models

import (
	"logtheus/internal/consts"
	"time"
)

type Token struct {
	ID        uint             `gorm:"primarykey"`
	CreatedAt time.Time        `gorm:"not null"`
	UserID    uint             `gorm:"not null;index"`
	User      User             `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE;"`
	Token     string           `gorm:"uniqueIndex;not null"`
	Type      consts.TokenType `gorm:"not null"`
	ExpiresAt time.Time        `gorm:"not null"`
}
