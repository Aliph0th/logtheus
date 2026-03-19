package dto

type ApplicationCreateRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	ProjectID   uint64 `json:"project_id" binding:"required"`
}

type ApplicationUpdateRequest struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
}

type ApplicationDTO struct {
	Id          uint64 `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	ProjectId   uint64 `json:"project_id"`
	CreatedAt   int64  `json:"created_at"`
	UpdatedAt   int64  `json:"updated_at"`
}
