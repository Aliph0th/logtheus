package models

import "gorm.io/gorm"

type User struct {
	Model           gorm.Model `gorm:"embedded"`
	Email           string     `gorm:"uniqueIndex;not null;"`
	Password        string     `gorm:"not null"`
	Username        string     `gorm:"not null"`
	IsEmailVerified bool       `gorm:"default:false"`
}
