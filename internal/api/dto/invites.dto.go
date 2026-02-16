package dto

import "logtheus/internal/consts/enums"

type InviteCreateRequest struct {
	ProjectID uint64            `json:"projectID" binding:"required"`
	Email     string            `json:"email" binding:"required"`
	Role      enums.ProjectRole `json:"role" binding:"required"`
}
