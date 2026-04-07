package models

import (
	"time"

	"github.com/google/uuid"
)

type InviteToken struct {
	Token        uuid.UUID  `gorm:"type:uuid;primaryKey" json:"token"`
	InviteeEmail string     `gorm:"primaryKey;autoIncrement:false" json:"invitee_email"`
	ProjectID    uint64     `gorm:"not null" json:"project_id"`
	Project      Project    `gorm:"foreignKey:ProjectID;constraint:OnDelete:CASCADE;" json:"-"`
	ExpiresAt    *time.Time `json:"expires_at"`

	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
}
