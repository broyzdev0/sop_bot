package utils

import (
	"sopbot/config"
	"strings"
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
func EscapeMarkdown(text string) string {
	replacer := strings.NewReplacer(
		"_", "\\_",
		"*", "\\*",
		"[", "\\[",
		"]", "\\]",
		"(", "\\(",
		")", "\\)",
		"~", "\\~",
		"`", "\\`",
		">", "\\>",
		"#", "\\#",
		"+", "\\+",
		"-", "\\-",
		"=", "\\=",
		"|", "\\|",
		"{", "\\{",
		"}", "\\}",
		".", "\\.",
		"!", "\\!",
	)
	return replacer.Replace(text)
}

func SplitMessage(msg string, maxLen int) []string {
	var parts []string
	for len(msg) > maxLen {
		splitAt := maxLen
		// Cari spasi terdekat ke belakang biar gak putus kata
		for i := maxLen; i > 0; i-- {
			if msg[i] == ' ' {
				splitAt = i
				break
			}
		}
		parts = append(parts, msg[:splitAt])
		msg = msg[splitAt:]
	}
	if len(msg) > 0 {
		parts = append(parts, msg)
	}
	return parts
}
