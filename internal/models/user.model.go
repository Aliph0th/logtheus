package models

import (
	"time"
)

type User struct {
	ID              uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	Email           string    `gorm:"uniqueIndex;not null;" json:"email"`
	Password        string    `gorm:"not null" json:"-"`
	Username        string    `gorm:"not null" json:"username"`
	IsEmailVerified bool      `gorm:"default:false" json:"isEmailVerified"`
	CreatedAt       time.Time `gorm:"not null" json:"createdAt"`
	UpdatedAt       time.Time `gorm:"not null" json:"updatedAt"`
}
