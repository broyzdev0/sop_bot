package data

import (
	"context"
	"fmt"
	"os"
	"strings"
	"log"
	"sopbot/config"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)
var (
    infoLog  = log.New(os.Stdout, "[INFO] ", log.LstdFlags)
    errorLog = log.New(os.Stderr, "[ERROR] ", log.LstdFlags)
	
)


func GetSOPFromMessage(message string) string {
	ctx := context.Background()

	b, err := os.ReadFile("credentials.json")
	if err != nil {
		errorLog.Printf("Gagal baca credentials: %v", err)
		return "❌ Gagal baca credentials"
	}

	conf, err := google.JWTConfigFromJSON(b, sheets.SpreadsheetsReadonlyScope)
	if err != nil {
		errorLog.Printf("Gagal parse JWT: %v", err)
		return fmt.Sprintf("❌ Gagal parse JWT: %v", err)
	}
	client := conf.Client(ctx)

	srv, err := sheets.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		errorLog.Printf("Gagal buat service: %v", err)
		return fmt.Sprintf("❌ Gagal buat service: %v", err)
	}

	resp, err := srv.Spreadsheets.Values.Get(config.SpreadsheetID, config.SheetRange).Do()
	if err != nil {
		errorLog.Printf("Gagal akses Google Sheet: %v", err)
		return fmt.Sprintf("❌ Gagal akses Google Sheet: %v", err)
	}

	lowerMessage := strings.TrimSpace(strings.ToLower(message))

	for _, row := range resp.Values {
		if len(row) >= 2 {
			keyword := strings.ToLower(strings.TrimSpace(row[0].(string)))
			desc := strings.TrimSpace(row[1].(string))

			// Hanya cocok jika sama persis
			if lowerMessage == keyword {
				return desc
			}
		}
	}

	infoLog.Printf("SOP tidak ditemukan (exact match only): %v", message)
	return "❓ SOP tidak ditemukan. Pastikan keyword-nya sesuai atau ketik /help untuk melihat daftar keyword, /tanya menggunakan AI, /ask untuk bertanya sop."
}


func GetAllKeywords() string {
	ctx := context.Background()

	b, err := os.ReadFile("credentials.json")
	if err != nil {
		errorLog.Printf("Gagal baca credentials: %v", err)
		return "❌ Gagal baca credentials"
	}

	conf, err := google.JWTConfigFromJSON(b, sheets.SpreadsheetsReadonlyScope)
	if err != nil {
		errorLog.Printf("Gagal parse JWT: %v", err)
		return "❌ Gagal parse credentials"
	}
	client := conf.Client(ctx)

	srv, err := sheets.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		errorLog.Printf("Gagal buat service Google Sheet: %v", err)
		return "❌ Gagal buat service Google Sheet"
	}

	resp, err := srv.Spreadsheets.Values.Get(config.SpreadsheetID, config.SheetRange).Do()
	if err != nil {
		errorLog.Printf("Gagal akses Google Sheet: %v", err)
		return "❌ Gagal akses Google Sheet"
	}

	var keywords []string
	for _, row := range resp.Values {
		if len(row) >= 1 {
			keyword := strings.TrimSpace(row[0].(string))
			if keyword != "" {
				keywords = append(keywords, "• " + keyword)
			}
		}
	}

	if len(keywords) == 0 {
		infoLog.Println("Belum ada keyword SOP yang tersedia.")
		return "❓ Belum ada keyword SOP yang tersedia."
	}

return "📚 *Daftar Keyword SOP:*\n\n" + strings.Join(keywords, "\n") + "\n\nKetik aja salah satu keyword di atas, nanti aku bantuin kasih SOP-nya ya!\n\nNote: /help untuk melihat daftar keyword, /tanya menggunakan AI, /ask untuk bertanya sop."

}
