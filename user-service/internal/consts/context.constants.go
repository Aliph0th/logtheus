package consts

const AUTH_CONTEXT_KEY = "user_auth_data"

var PUBLIC_METHODS = map[string]bool{
	"/logtheus.user.v1.UserService/RegisterUser": true,
	"/logtheus.user.v1.UserService/LoginUser":    true,
	// "/user.v1.UserService/ResetPassword":   true,
}
