package types

type VerifyEmailData struct {
	Username  string
	Url       string
	Code      string
	ExpiresIn uint8
}

type InviteEmailData struct {
	InviterName string
	ProjectName string
	InviteLink  string
}
