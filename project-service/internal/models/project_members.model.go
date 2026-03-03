package models

import (
	"logtheus/shared/pkg/consts"
	"time"
)

type ProjectMember struct {
	ProjectID uint64  `gorm:"primaryKey;autoIncrement:false" json:"projectID"`
	UserID    uint64  `gorm:"primaryKey;autoIncrement:false" json:"userID"`
	InvitedBy *uint64 `json:"invitedBy"`
	Project   Project `gorm:"foreignKey:ProjectID;constraint:OnDelete:CASCADE;" json:"-"`

	Role consts.ProjectRole `gorm:"not null;default:'viewer'" json:"role"`

	JoinedAt  *time.Time `json:"joinedAt"`
	CreatedAt time.Time  `gorm:"autoCreateTime;nano" json:"createdAt"`
}
