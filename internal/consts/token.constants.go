package consts

import "time"

const (
	ACCESS_TOKEN_TTL  = time.Hour
	REFRESH_TOKEN_TTL = time.Hour * 24 * 7
	VERIFY_TOKEN_TTL  = time.Minute * 15
	RESET_TOKEN_TTL   = time.Minute * 5

	VERIFY_TOKEN_LENGTH = 6
)
