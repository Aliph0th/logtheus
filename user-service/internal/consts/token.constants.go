package consts

import "time"

type TokenType string

const (
	TOKEN_TYPE_REFRESH        TokenType = "refresh_token"
	TOKEN_TYPE_VERIFY         TokenType = "verify_token"
	TOKEN_TYPE_PASSWORD_RESET TokenType = "reset_token"
)

const (
	TTL_ACCESS_TOKEN  = time.Hour
	TTL_REFRESH_TOKEN = time.Hour * 24 * 7
	TTL_VERIFY_TOKEN  = time.Minute * 15
	TTL_RESET_TOKEN   = time.Minute * 5

	VERIFY_EMAIL_TOKEN_LENGTH = 6
)
