package data

import "sync"

var (
	userHistories = make(map[int64][]string)
	mu            sync.Mutex
	maxHistory    = 10
)

func AppendToHistory(userID int64, message string) {
	mu.Lock()
	defer mu.Unlock()

	userHistories[userID] = append(userHistories[userID], message)

	if len(userHistories[userID]) > maxHistory {
		userHistories[userID] = userHistories[userID][1:]
	}
}

func GetHistory(userID int64) []string {
	mu.Lock()
	defer mu.Unlock()

	return append([]string(nil), userHistories[userID]...)
}

func ResetHistory(userID int64) {
	mu.Lock()
	defer mu.Unlock()
	userHistories[userID] = []string{}
}
