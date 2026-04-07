package models

import (
	"time"
)

type User struct {
	ID              uint64    `gorm:"primaryKey" json:"id"`
	Email           string    `gorm:"uniqueIndex;not null;" json:"email"`
	Password        string    `gorm:"not null" json:"-"`
	Username        string    `gorm:"not null" json:"username"`
	IsEmailVerified bool      `gorm:"default:false" json:"is_email_verified"`
	CreatedAt       time.Time `gorm:"not null" json:"created_at"`
	UpdatedAt       time.Time `gorm:"not null" json:"updated_at"`
}
