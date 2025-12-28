package dto

type ProjectCreateRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
}
