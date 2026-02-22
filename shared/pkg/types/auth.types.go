package types

import "github.com/golang-jwt/jwt/v5"

type UserAuthClaims struct {
	jwt.RegisteredClaims
	UserAuthPayload
}

type UserAuthPayload struct {
	UserID          uint64 `json:"userID,omitempty"`
	IsEmailVerified bool   `json:"isEmailVerified"`
}
