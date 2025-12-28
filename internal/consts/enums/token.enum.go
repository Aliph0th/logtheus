package enums

type TokenType string

const (
	TOKEN_TYPE_REFRESH        TokenType = "refresh_token"
	TOKEN_TYPE_VERIFY         TokenType = "verify_token"
	TOKEN_TYPE_PASSWORD_RESET TokenType = "reset_token"
)
