package models

import (
	"logtheus/internal/consts/enums"
	"time"
)

type Token struct {
	Token     string          `gorm:"primaryKey;autoIncrement:false"`
	UserID    uint64          `gorm:"not null;index"`
	User      User            `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE;"`
	Type      enums.TokenType `gorm:"not null"`
	ExpiresAt time.Time       `gorm:"not null"`
	CreatedAt time.Time       `gorm:"not null"`
}
