package main

import "sopbot/bot"

func main() {
	bot.StartBot()
}

func splitMessage(text string, limit int) []string {
	if len(text) <= limit {
		return []string{text}
	}
	var parts []string
	for len(text) > limit {
		part := text[:limit]
		text = text[limit:]
		parts = append(parts, part)
	}
	if len(text) > 0 {
		parts = append(parts, text)
	}
	return parts
}
