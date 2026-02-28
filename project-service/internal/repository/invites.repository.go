package repository

import (
	"logtheus/shared/pkg/storages"

	"gorm.io/gorm"
)

type InvitesRepository struct {
	db *gorm.DB
}

func NewInvitesRepository(db *storages.Database) *InvitesRepository {
	return &InvitesRepository{db.DB}
}
