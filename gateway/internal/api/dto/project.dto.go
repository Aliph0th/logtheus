package dto

import "logtheus/shared/pkg/consts"

type ProjectCreateRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
}

type ProjectUpdateRequest struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
}

type ProjectDTO struct {
	Id          uint64 `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	CreatedAt   int64  `json:"created_at"`
	UpdatedAt   int64  `json:"updated_at"`
}

type ProjectMemberUserDTO struct {
	Id       uint64 `json:"id"`
	Email    string `json:"email"`
	Username string `json:"username"`
}

type ProjectMemberDTO struct {
	UserId   uint64             `json:"user_id"`
	Role     consts.ProjectRole `json:"role"`
	JoinedAt *int64             `json:"joined_at"`
	User     *ProjectMemberUserDTO `json:"user"`
}
