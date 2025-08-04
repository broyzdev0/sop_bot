package utils

import "time"

var UserLastMessage = make(map[int]int64)

const DelaySeconds = 0

func IsSpam(userID int) (bool, string) {
	now := time.Now().Unix()
	if lastTime, ok := UserLastMessage[userID]; ok {
		if now-lastTime < DelaySeconds {
			return true, "Sabar bang, delay 5 detik dulu ya 😅"
		}
	}
	UserLastMessage[userID] = now

	return false, ""
}
