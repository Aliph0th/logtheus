package types

import "time"

type VerifyEmailEvent struct {
	Username          string `json:"username"`
	Email             string `json:"email"`
	Code              string `json:"code"`
	ExpirationMinutes uint8  `json:"expiration_minutes"`
}

type InviteEmailEvent struct {
	InviteeName string    `json:"invitee_name"`
	Referrer    string    `json:"referrer"`
	ProjectName string    `json:"project_name"`
	Email       string    `json:"email"`
	Code        string    `json:"code"`
	ExpiresAt   time.Time `json:"expires_at"`
}
