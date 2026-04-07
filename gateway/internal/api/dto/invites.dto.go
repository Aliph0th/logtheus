package dto

import "logtheus/shared/pkg/consts"

type InviteCreateRequest struct {
	ProjectID uint64             `json:"project_id" binding:"required"`
	Email     string             `json:"email" binding:"required"`
	Role      consts.ProjectRole `json:"role" binding:"required"`
	ExpiresAt *string            `json:"expires_at,omitempty"`
}
