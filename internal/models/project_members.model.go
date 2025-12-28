package models

import (
	"database/sql"
	"logtheus/internal/consts/enums"
	"time"
)

type ProjectMember struct {
	ProjectID uint64  `gorm:"primaryKey;autoIncrement:false" json:"projectID"`
	UserID    uint64  `gorm:"primaryKey;autoIncrement:false" json:"userID"`
	InvitedBy uint64  `gorm:"not null" json:"invitedBy"`
	Project   Project `gorm:"foreignKey:ProjectID;constraint:OnDelete:CASCADE;" json:"-"`
	User      User    `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE;" json:"-"`
	Inviter   User    `gorm:"foreignKey:InvitedBy;constraint:OnDelete:CASCADE;" json:"-"`

	Role       enums.ProjectRole `gorm:"not null;default:'viewer'" json:"role"`
	IsAccepted bool              `gorm:"not null;default:false" json:"isAccepted"`

	JoinedAt  sql.NullTime `json:"joinedAt"`
	CreatedAt time.Time    `gorm:"autoCreateTime;nano" json:"createdAt"`
}
