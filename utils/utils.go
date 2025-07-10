package utils

import (
	"sopbot/config"
	
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func isAdmin(userID int64) bool {
	for _, id := range config.AdminUserIDs {
		if userID == id {
			return true
		}
	}
	return false
}

func WaitNextMessage(updates tgbotapi.UpdatesChannel, userID int64) string {
	for update := range updates {
		if update.Message != nil && update.Message.From.ID == userID {
			return update.Message.Text
		}
	}
	return ""
}
