package models

import (
	"time"
)

type User struct {
	ID              uint      `gorm:"primaryKey;autoIncrement"`
	Email           string    `gorm:"uniqueIndex;not null;"`
	Password        string    `gorm:"not null"`
	Username        string    `gorm:"not null"`
	IsEmailVerified bool      `gorm:"default:false"`
	CreatedAt       time.Time `gorm:"not null"`
	UpdatedAt       time.Time `gorm:"not null"`
}
