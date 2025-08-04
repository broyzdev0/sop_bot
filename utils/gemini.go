package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"sopbot/config"
)

var GeminiAPIKey = config.GeminiAPIKey

type geminiRequest struct {
	Contents []geminiContent `json:"contents"`
}

type geminiContent struct {
	Parts []geminiPart `json:"parts"`
	Role  string       `json:"role"`
}

type geminiPart struct {
	Text string `json:"text"`
}

type geminiResponse struct {
	Candidates []struct {
		Content struct {
			Parts []struct {
				Text string `json:"text"`
			} `json:"parts"`
		} `json:"content"`
	} `json:"candidates"`
}

func AskGemini(question string, history []string) (string, error) {
	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/gemini-1.5-flash-latest:generateContent?key=%s", GeminiAPIKey)

	var contents []geminiContent

	// Tambahkan semua history ke dalam isi request
	for _, msg := range history {
		role := "user"
		if len(msg) >= 5 && msg[:5] == "Bot: " {
			role = "model"
			msg = msg[5:]
		} else if len(msg) >= 6 && msg[:6] == "User: " {
			msg = msg[6:]
		}

		contents = append(contents, geminiContent{
			Role:  role,
			Parts: []geminiPart{{Text: msg}},
		})
	}

	// Tambahkan pertanyaan terakhir
	contents = append(contents, geminiContent{
		Role:  "user",
		Parts: []geminiPart{{Text: "Jawab dalam Bahasa Indonesia: " + question}},
	})

	requestBody := geminiRequest{
		Contents: contents,
	}

	jsonData, _ := json.Marshal(requestBody)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	// fmt.Println("DEBUG Gemini raw response:", string(body))

	var result geminiResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return "", err
	}

	if len(result.Candidates) > 0 && len(result.Candidates[0].Content.Parts) > 0 {
		return result.Candidates[0].Content.Parts[0].Text, nil
	}
	return "❌ Tidak ada respons dari Gemini.", nil
}
