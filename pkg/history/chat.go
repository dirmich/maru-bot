package history

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// Message represents a single chat message
type Message struct {
	ID        string `json:"id"`
	Role      string `json:"role"` // "user" or "assistant"
	Content   string `json:"content"`
	Timestamp string `json:"timestamp"`
}

// ChatHistoryManager handles saving and loading chat logs
type ChatHistoryManager struct {
	historyDir string
}

// NewChatHistoryManager creates a new history manager
func NewChatHistoryManager(baseDir string) *ChatHistoryManager {
	dir := filepath.Join(baseDir, "history", "chat")
	return &ChatHistoryManager{
		historyDir: dir,
	}
}

// SaveMessage appends a message to the current day's history file
func (m *ChatHistoryManager) SaveMessage(msg Message) error {
	if err := os.MkdirAll(m.historyDir, 0755); err != nil {
		return fmt.Errorf("failed to create history dir: %w", err)
	}

	date := time.Now().Format("2006-01-02")
	filename := filepath.Join(m.historyDir, date+".json")

	var messages []Message
	data, err := os.ReadFile(filename)
	if err == nil {
		if err := json.Unmarshal(data, &messages); err != nil {
			// If file is corrupted, start fresh
			messages = []Message{}
		}
	}

	messages = append(messages, msg)
	
	newData, err := json.MarshalIndent(messages, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal messages: %w", err)
	}

	return os.WriteFile(filename, newData, 0644)
}

// GetHistory returns the messages for a specific date (YYYY-MM-DD)
func (m *ChatHistoryManager) GetHistory(date string) ([]Message, error) {
	filename := filepath.Join(m.historyDir, date+".json")
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var messages []Message
	if err := json.Unmarshal(data, &messages); err != nil {
		return nil, err
	}

	return messages, nil
}

// ListDays returns a list of dates that have chat history
func (m *ChatHistoryManager) ListDays() ([]string, error) {
	if _, err := os.Stat(m.historyDir); os.IsNotExist(err) {
		return []string{}, nil
	}

	files, err := os.ReadDir(m.historyDir)
	if err != nil {
		return nil, err
	}

	var days []string
	for _, f := range files {
		if !f.IsDir() && strings.HasSuffix(f.Name(), ".json") {
			days = append(days, strings.TrimSuffix(f.Name(), ".json"))
		}
	}

	// Sort newest first
	sort.Slice(days, func(i, j int) bool {
		return days[i] > days[j]
	})

	return days, nil
}

// DeleteHistory removes history for a specific date
func (m *ChatHistoryManager) DeleteHistory(date string) error {
	filename := filepath.Join(m.historyDir, date+".json")
	return os.Remove(filename)
}
