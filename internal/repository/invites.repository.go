package repository

import "gorm.io/gorm"

type InviteRepository struct {
	db *gorm.DB
}

func NewInvitesRepository(db *gorm.DB) *InviteRepository {
	return &InviteRepository{db}
}
