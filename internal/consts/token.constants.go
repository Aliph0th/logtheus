package consts

import "time"

type TokenType string

const (
	TYPE_REFRESH_TOKEN TokenType = "refresh_token"
	TYPE_VERIFY_TOKEN  TokenType = "verify_token"
	TYPE_RESET_TOKEN   TokenType = "reset_token"
)

const (
	ACCESS_TOKEN_TTL  = time.Hour
	REFRESH_TOKEN_TTL = time.Hour * 24 * 7
	VERIFY_TOKEN_TTL  = time.Minute * 15
	RESET_TOKEN_TTL   = time.Minute * 5

	VERIFY_TOKEN_LENGTH = 6
)
