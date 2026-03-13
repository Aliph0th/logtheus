package models

import (
	"logtheus/shared/pkg/consts"
	"time"
)

type ProjectMember struct {
	ProjectID uint64  `gorm:"primaryKey;autoIncrement:false" json:"project_id"`
	UserID    uint64  `gorm:"primaryKey;autoIncrement:false" json:"user_id"`
	InvitedBy *uint64 `json:"invited_by"`
	Project   Project `gorm:"foreignKey:ProjectID;constraint:OnDelete:CASCADE;" json:"-"`

	Role consts.ProjectRole `gorm:"not null;default:'viewer'" json:"role"`

	JoinedAt  *time.Time `json:"joined_at"`
	CreatedAt time.Time  `gorm:"autoCreateTime;nano" json:"created_at"`
}
