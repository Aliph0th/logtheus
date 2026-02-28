package models

import (
	"database/sql"
	"logtheus/project/internal/consts"
	"time"
)

type ProjectMember struct {
	ProjectID uint64  `gorm:"primaryKey;autoIncrement:false" json:"projectID"`
	UserID    uint64  `gorm:"primaryKey;autoIncrement:false" json:"userID"`
	InvitedBy uint64  `gorm:"not null" json:"invitedBy"`
	Project   Project `gorm:"foreignKey:ProjectID;constraint:OnDelete:CASCADE;" json:"-"`

	Role consts.ProjectRole `gorm:"not null;default:'viewer'" json:"role"`

	JoinedAt  sql.NullTime `json:"joinedAt"`
	CreatedAt time.Time    `gorm:"autoCreateTime;nano" json:"createdAt"`
}
