package consts

import "fmt"

func REDIS_KEY_EMAIL_VERIFICATION(userID uint64) string {
	return fmt.Sprintf("email_verification:%d", userID)
}
