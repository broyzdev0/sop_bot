package bot

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sopbot/config"
	"sopbot/data"
	"sopbot/utils"
	"strings"
	"syscall"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	gOAuth "golang.org/x/oauth2/google"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

var (
	infoLog  = log.New(os.Stdout, "[INFO] ", log.LstdFlags)
	errorLog = log.New(os.Stderr, "[ERROR] ", log.LstdFlags)
)

func StartBot() {
	bot, err := tgbotapi.NewBotAPI(config.BotToken)
	if err != nil {
		log.Fatal(err)
	}
	infoLog.Printf("✅ Bot Run Success")
	// Kirim pesan bot ON
	// msg1 := tgbotapi.NewMessage(config.AllowedGroupID, "✅ Bot aktif (ON)")
	msg1 := tgbotapi.NewMessage(config.AllowedGroupID, "⚠️ Maintenance Mode")
	msg2 := tgbotapi.NewMessage(config.AllowedGroupID, "📢 Ketik /help Untuk Memulai")
	msg3 := tgbotapi.NewMessage(config.AllowedGroupID, "✅ Created By Broyzdev 2025")
	bot.Send(msg1)
	bot.Send(msg2)
	bot.Send(msg3)

	// Tangani sinyal keluar
	go func() {
		sigs := make(chan os.Signal, 1)
		signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

		<-sigs
		infoLog.Printf("⚠️ Bot dimatikan (OFF)")
		exitMsg := tgbotapi.NewMessage(config.AllowedGroupID, "⚠️ Bot dimatikan (OFF)")
		bot.Send(exitMsg)
		os.Exit(0)
	}()

	// 🧹 Hapus webhook sebelum menggunakan polling
	_, err = bot.Request(tgbotapi.DeleteWebhookConfig{})
	if err != nil {
		errorLog.Printf("❌ Gagal menghapus webhook: %v", err)
	}
	infoLog.Println("✅ Webhook dihapus")

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil {
			continue
		}

		chat := update.Message.Chat
		if chat.IsPrivate() || chat.ID != config.AllowedGroupID {
			continue
		}

		userID := update.Message.From.ID
		chatID := chat.ID
		query := update.Message.Text
		user := update.Message.From
		lowerText := strings.ToLower(query)

		infoLog.Printf("Pesan dari: %s | UserID: %d | ChatID: %d", user.UserName, userID, chatID)

		// Jawaban otomatis khusus
		if strings.Contains(lowerText, "angga") {
			reply := tgbotapi.NewMessage(chatID, "Angga Nur Gunawan adalah penciptaku.")
			bot.Send(reply)
			continue
		}

		if strings.Contains(lowerText, "siapa yang membuat") ||
			strings.Contains(lowerText, "pembuatmu") ||
			strings.Contains(lowerText, "dibuat oleh siapa") ||
			strings.Contains(lowerText, "pembuat kamu") ||
			strings.Contains(lowerText, "siapa penciptamu") {
			reply := tgbotapi.NewMessage(chatID, "Saya dibuat oleh Angga Nur Gunawan pada tahun 2025.")
			bot.Send(reply)
			continue
		}

		// 🔁 Auto-reply jika user membalas pesan dari bot
		if update.Message.ReplyToMessage != nil {
			original := update.Message.ReplyToMessage
			if original.From != nil && original.From.UserName == bot.Self.UserName {
				isSpam, spamMsg := utils.IsSpam(int(userID))
				if isSpam {
					bot.Send(tgbotapi.NewMessage(chatID, spamMsg))
					continue
				}

				question := update.Message.Text
				infoLog.Printf("[AutoReplyAI] %s membalas pesan bot: %s", user.UserName, question)

				data.AppendToHistory(userID, "User: "+question)
				history := data.GetHistory(userID)

				answer, err := utils.AskGemini(question, history)
				if err != nil {
					errorLog.Printf("❌ Gagal tanya ke Gemini (auto-reply): %v", err)
					bot.Send(tgbotapi.NewMessage(chatID, "❌ Gagal mendapatkan jawaban dari AI."))
					continue
				}

				data.AppendToHistory(userID, "Bot: "+answer)

				parts := utils.SplitMessage(answer, 4000)
				for _, part := range parts {
					escaped := utils.EscapeMarkdown(part)
					msg := tgbotapi.NewMessage(chatID, escaped)
					msg.ParseMode = "MarkdownV2"
					msg.ReplyToMessageID = update.Message.MessageID
					bot.Send(msg)
				}
				continue
			}
		}

		if strings.HasPrefix(query, "/help") {
			keywords := data.GetAllKeywords()
			msg := tgbotapi.NewMessage(chatID, keywords)
			msg.ParseMode = "Markdown"
			bot.Send(msg)
			continue
		}

		if isAdmin(userID) {
			if strings.HasPrefix(query, "/add ") {
				args := strings.SplitN(strings.TrimPrefix(query, "/add "), "|", 2)
				if len(args) == 2 {
					keyword := strings.TrimSpace(args[0])
					newDesc := strings.TrimSpace(args[1])
					infoLog.Printf("/add command by %s: keyword='%s', newDesc='%s'", user.UserName, keyword, newDesc)
					msg := tgbotapi.NewMessage(chatID, AddSOP(keyword, newDesc))
					bot.Send(msg)
				} else {
					bot.Send(tgbotapi.NewMessage(chatID, "❌ Format salah. Contoh: /add keyword | deskripsi"))
				}
				continue
			}

			if strings.HasPrefix(query, "/delete ") {
				keyword := strings.TrimSpace(strings.TrimPrefix(query, "/delete "))
				infoLog.Printf("/delete command by %s: keyword='%s'", user.UserName, keyword)
				msg := tgbotapi.NewMessage(chatID, DeleteSOP(keyword))
				bot.Send(msg)
				continue
			}

			if strings.HasPrefix(query, "/edit ") {
				args := strings.SplitN(strings.TrimPrefix(query, "/edit "), "|", 2)
				if len(args) == 2 {
					keyword := strings.TrimSpace(args[0])
					newDesc := strings.TrimSpace(args[1])
					infoLog.Printf("/edit command by %s: keyword='%s', newDesc='%s'", user.UserName, keyword, newDesc)
					msg := tgbotapi.NewMessage(chatID, EditSOP(keyword, newDesc))
					bot.Send(msg)
				} else {
					bot.Send(tgbotapi.NewMessage(chatID, "❌ Format salah. Contoh: /edit keyword | deskripsi baru"))
				}
				continue
			}

			if query == "/reset" {
				data.ResetHistory(userID)
				bot.Send(tgbotapi.NewMessage(chatID, "✅ Riwayat percakapanmu sudah direset."))
				continue
			}
		}

		if strings.HasPrefix(query, "/ask ") {
			isSpam, spamMsg := utils.IsSpam(int(userID))
			if isSpam {
				msg := tgbotapi.NewMessage(chatID, spamMsg)
				bot.Send(msg)
				continue
			}
			queryContent := strings.TrimPrefix(query, "/ask ")
			infoLog.Printf("/ask command: %s | From: %s", queryContent, user.UserName)
			result := data.GetSOPFromMessage(queryContent)
			msg := tgbotapi.NewMessage(chatID, result)
			bot.Send(msg)

		}

		if strings.HasPrefix(query, "/tanya ") {
			isSpam, spamMsg := utils.IsSpam(int(userID))
			if isSpam {
				msg := tgbotapi.NewMessage(chatID, spamMsg)
				bot.Send(msg)
				continue
			}

			question := strings.TrimPrefix(query, "/tanya ")
			infoLog.Printf("/tanya command: %s | From: %s", question, user.UserName)

			// Simpan pertanyaan ke history user
			data.AppendToHistory(userID, "User: "+question)

			// Ambil history percakapan
			history := data.GetHistory(userID)

			answer, err := utils.AskGemini(question, history)
			if err != nil {
				errorLog.Printf("Gagal tanya ke Gemini: %v", err)
				bot.Send(tgbotapi.NewMessage(chatID, "❌ Gagal tanya ke Gemini"))
				continue
			}

			// Simpan jawaban ke history juga
			data.AppendToHistory(userID, "Bot: "+answer)

			// Escape markdown & kirim bertahap
			parts := utils.SplitMessage(answer, 4000)
			for _, part := range parts {
				escaped := utils.EscapeMarkdown(part)
				msg := tgbotapi.NewMessage(chatID, escaped)
				msg.ParseMode = "MarkdownV2"
				bot.Send(msg)
			}
		}

	}
}

func isAdmin(userID int64) bool {
	for _, id := range config.AdminUserIDs {
		if userID == id {
			return true
		}
	}
	return false
}

func getSheetsService(ctx context.Context) (*sheets.Service, error) {
	credentialsJSON, err := os.ReadFile("credentials.json")
	if err != nil {
		errorLog.Printf("Gagal membaca file credentials.json: %v", err)
		return nil, fmt.Errorf("gagal membaca file credentials.json: %w", err)
	}
	conf, err := gOAuth.JWTConfigFromJSON(credentialsJSON, sheets.SpreadsheetsScope)
	if err != nil {
		errorLog.Printf("JWTConfigFromJSON error: %v", err)
		return nil, fmt.Errorf("JWTConfigFromJSON error: %w", err)
	}
	client := conf.Client(ctx)
	return sheets.NewService(ctx, option.WithHTTPClient(client))
}

func AddSOP(keyword, description string) string {
	ctx := context.Background()
	srv, err := getSheetsService(ctx)
	if err != nil {
		log.Println("Gagal inisialisasi Google Sheets API:", err)
		return "❌ Gagal inisialisasi Google Sheets API"
	}

	// Cek apakah keyword sudah ada
	resp, err := srv.Spreadsheets.Values.Get(config.SpreadsheetID, config.SheetRange).Do()
	if err != nil {
		log.Println("Gagal akses Google Sheet:", err)
		return "❌ Gagal akses Google Sheet"
	}

	for _, row := range resp.Values {
		if len(row) > 0 && strings.EqualFold(row[0].(string), keyword) {
			log.Println("Keyword sudah ada:", keyword)
			return "⚠️ Keyword sudah ada, tidak disimpan ulang."
		}
	}

	// Kalau belum ada, tambahkan
	newRow := []interface{}{keyword, description}
	appendCall := &sheets.ValueRange{Values: [][]interface{}{newRow}}
	_, err = srv.Spreadsheets.Values.Append(config.SpreadsheetID, config.SheetRange, appendCall).ValueInputOption("RAW").Do()
	if err != nil {
		log.Println("Gagal menambahkan SOP ke Sheet:", err)
		return "❌ Gagal menambahkan SOP"
	}
	return "✅ SOP berhasil ditambahkan!"
}

func DeleteSOP(keyword string) string {
	ctx := context.Background()
	srv, err := getSheetsService(ctx)
	if err != nil {
		log.Println("Gagal inisialisasi Google Sheets API:", err)
		return "❌ Gagal inisialisasi Google Sheets API"
	}
	resp, err := srv.Spreadsheets.Values.Get(config.SpreadsheetID, config.SheetRange).Do()
	if err != nil {
		log.Println("Gagal akses Google Sheet:", err)
		return "❌ Gagal akses Google Sheet"
	}
	var matchedRow int = -1
	for i, row := range resp.Values {
		if len(row) > 0 && strings.EqualFold(row[0].(string), keyword) {
			matchedRow = i
			break
		}
	}
	if matchedRow == -1 {
		infoLog.Println("SOP tidak ditemukan")
		return "❌ SOP tidak ditemukan"
	}
	request := &sheets.BatchUpdateSpreadsheetRequest{
		Requests: []*sheets.Request{{
			DeleteDimension: &sheets.DeleteDimensionRequest{
				Range: &sheets.DimensionRange{
					SheetId:    0,
					Dimension:  "ROWS",
					StartIndex: int64(matchedRow),
					EndIndex:   int64(matchedRow + 1),
				},
			},
		}},
	}
	_, err = srv.Spreadsheets.BatchUpdate(config.SpreadsheetID, request).Do()
	if err != nil {
		infoLog.Printf("Gagal menghapus SOP: %v", err)
		return "❌ Gagal menghapus SOP"
	}
	infoLog.Println("SOP berhasil dihapus!")
	return "✅ SOP berhasil dihapus!"
}

func EditSOP(keyword, newDescription string) string {

	ctx := context.Background()
	srv, err := getSheetsService(ctx)
	if err != nil {
		errorLog.Printf("Gagal inisialisasi Google Sheets API: %v", err)
		return "❌ Gagal inisialisasi Google Sheets API"
	}
	resp, err := srv.Spreadsheets.Values.Get(config.SpreadsheetID, config.SheetRange).Do()
	if err != nil {
		errorLog.Printf("Gagal akses Google Sheet: %v", err)
		return "❌ Gagal akses Google Sheet"
	}
	var matchedRow int = -1
	for i, row := range resp.Values {
		if len(row) > 0 && strings.EqualFold(row[0].(string), keyword) {
			matchedRow = i
			break
		}
	}
	if matchedRow == -1 {
		errorLog.Printf("SOP tidak ditemukan: %v", err)
		return "❌ SOP tidak ditemukan"
	}
	rangeToUpdate := fmt.Sprintf("Sheet1!B%d", matchedRow+1)
	_, err = srv.Spreadsheets.Values.Update(config.SpreadsheetID, rangeToUpdate, &sheets.ValueRange{
		Values: [][]interface{}{{newDescription}},
	}).ValueInputOption("RAW").Do()
	if err != nil {
		errorLog.Printf("Gagal Update Sop: %v", err)
		return "❌ Gagal mengubah SOP"
	}
	infoLog.Println("SOP berhasil diubah!")
	return "✅ SOP berhasil diubah!"
}
