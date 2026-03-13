package repository

import (
	"errors"
	"logtheus/project/internal/models"
	"logtheus/shared/pkg/consts"
	"logtheus/shared/pkg/storages"

	"gorm.io/gorm"
)

type MemberRepository struct {
	db *gorm.DB
}

func NewMemberRepository(db *storages.Database) *MemberRepository {
	return &MemberRepository{db.DB}
}

func (r *MemberRepository) CountProjectMembers(projectID uint64) (uint8, error) {
	var count int64
	if err := r.db.Model(&models.ProjectMember{}).Where("project_id = ?", projectID).Count(&count).Error; err != nil {
		return 0, err
	}
	return uint8(count), nil
}

func (r *MemberRepository) GetMemberRole(userID, projectID uint64) (*consts.ProjectRole, error) {
	var member models.ProjectMember
	if err := r.db.First(&member, "user_id = ? AND project_id = ? AND joined_at IS NOT NULL", userID, projectID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &member.Role, nil
}

func (r *MemberRepository) AddMember(member *models.ProjectMember) error {
	return r.db.Create(member).Error
}

func (r *MemberRepository) SaveToken(token *models.InviteToken) error {
	return r.db.Create(token).Error
}

// FIXME:
func (r *MemberRepository) IsInvited(email string, userID *uint64, projectID uint64) (bool, error) {
	var exists bool
	var err error

	if userID == nil {
		err = r.db.Raw(`
			SELECT COUNT(*) > 0 FROM invite_tokens
			WHERE invitee_email = ? AND project_id = ?
		`, email, projectID).Scan(&exists).Error
	} else {
		err = r.db.Raw(`
			SELECT COUNT(*) > 0 FROM (
				SELECT 1 FROM invite_tokens
							WHERE invitee_email = ? AND project_id = ?
				UNION
				SELECT 1 FROM project_members
							WHERE user_id = ? AND project_id = ?
			) as T
		`, email, projectID, userID, projectID).Scan(&exists).Error
	}

	if err != nil {
		return false, err
	}
	return exists, nil
}
