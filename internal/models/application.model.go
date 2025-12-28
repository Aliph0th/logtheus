package models

type Application struct {
	ID          uint64  `gorm:"primaryKey" json:"id"`
	Name        string  `gorm:"not null" json:"name"`
	Description string  `json:"description"`
	ProjectID   uint64  `gorm:"not null" json:"projectID"`
	Project     Project `gorm:"foreignKey:ProjectID;constraint:OnDelete:CASCADE;" json:"project"`

	CreatedAt uint64 `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt uint64 `gorm:"autoUpdateTime" json:"updatedAt"`
}
