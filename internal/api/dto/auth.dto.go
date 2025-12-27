package dto

import "github.com/golang-jwt/jwt/v5"

type RegisterRequest struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
	Username string `json:"username" binding:"required"`
}

type VerifyEmailRequest struct {
	Code string `json:"code" binding:"required"`
}

type UserAuthClaims struct {
	jwt.RegisteredClaims
	UserAuthPayload
}

type UserAuthPayload struct {
	UserID          uint `json:"userID,omitempty"`
	IsEmailVerified bool `json:"isEmailVerified"`
}
