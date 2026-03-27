package consts

type LogFormat string

const (
	LOG_FORMAT_UNKNOWN LogFormat = "FORMAT_UNKNOWN"
	LOG_FORMAT_JSON    LogFormat = "FORMAT_JSON"
	LOG_FORMAT_NGINX   LogFormat = "FORMAT_NGINX"
	LOG_FORMAT_TEXT    LogFormat = "FORMAT_TEXT"
	LOG_FORMAT_BINARY  LogFormat = "FORMAT_BINARY"
)
