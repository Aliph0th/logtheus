package consts

import "time"

type TokenType string

const (
	REFRESH_TOKEN TokenType = "refresh_token"
	VERIFY_TOKEN  TokenType = "verify_token"
	RESET_TOKEN   TokenType = "reset_token"
)

const (
	ACCESS_TOKEN_TTL  = time.Hour
	REFRESH_TOKEN_TTL = time.Hour * 24 * 7

	VERIFY_TOKEN_LENGTH = 6
)
