package dto

import "github.com/golang-jwt/jwt/v5"

type RegisterRequest struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
	Username string `json:"username" binding:"required"`
}

type UserAuthClaims struct {
	jwt.RegisteredClaims
	UserID int `json:"userID,omitempty"`
}

type UserAuthPayload struct {
	UserID int `json:"userID,omitempty"`
}
