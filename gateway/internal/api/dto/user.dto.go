package dto

type UserDTO struct {
	Id              uint64 `json:"id"`
	Email           string `json:"email"`
	Username        string `json:"username"`
	IsEmailVerified bool   `json:"is_email_verified"`
	CreatedAt       int64  `json:"created_at"`
	UpdatedAt       int64  `json:"updated_at"`
}
