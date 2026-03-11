package session

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/dirmich/marubot/pkg/providers"
)

type Session struct {
	Key      string              `json:"key"`
	Messages []providers.Message `json:"messages"`
	Created  time.Time           `json:"created"`
	Updated  time.Time           `json:"updated"`
}

type SessionManager struct {
	storage string
	db      *SQLiteStore
}

func NewSessionManager(storage string) *SessionManager {
	dbPath := filepath.Join(storage, "history.db")
	db, err := NewSQLiteStore(dbPath)
	if err != nil {
		fmt.Printf("Warning: Failed to initialize SQLite session store: %v. Falling back to memory (non-persistent).\n", err)
	}

	sm := &SessionManager{
		storage: storage,
		db:      db,
	}

	return sm
}

func (sm *SessionManager) AddMessage(sessionKey, role, content string) {
	if sm.db != nil {
		if err := sm.db.SaveMessage(sessionKey, role, content); err != nil {
			fmt.Printf("Error saving message to SQLite: %v\n", err)
		}
	}
}

func (sm *SessionManager) GetHistory(key string) []providers.Message {
	if sm.db == nil {
		return []providers.Message{}
	}

	// For standard history, we return a reasonable amount (e.g. 50)
	msgs, err := sm.db.GetMessages(key, 50)
	if err != nil {
		fmt.Printf("Error getting history from SQLite: %v\n", err)
		return []providers.Message{}
	}
	return msgs
}

func (sm *SessionManager) SearchRelevant(query string, limit int) []providers.Message {
	if sm.db == nil {
		return nil
	}
	msgs, err := sm.db.SearchRelevant(query, limit)
	if err != nil {
		fmt.Printf("Error searching relevant messages: %v\n", err)
		return nil
	}
	return msgs
}

func (sm *SessionManager) Close() error {
	if sm.db != nil {
		return sm.db.Close()
	}
	return nil
}

// MigrateJSONToSQLite moves existing .json sessions to the sqlite database
func (sm *SessionManager) MigrateJSONToSQLite() error {
	if sm.db == nil || sm.storage == "" {
		return nil
	}

	files, err := os.ReadDir(sm.storage)
	if err != nil {
		return err
	}

	fmt.Printf("📦 Migrating session files from %s to SQLite...\n", sm.storage)
	count := 0
	for _, file := range files {
		if file.IsDir() || filepath.Ext(file.Name()) != ".json" {
			continue
		}

		path := filepath.Join(sm.storage, file.Name())
		data, err := os.ReadFile(path)
		if err != nil {
			fmt.Printf("  ⚠️ Failed to read %s: %v\n", file.Name(), err)
			continue
		}

		var session struct {
			Key      string              `json:"key"`
			Messages []providers.Message `json:"messages"`
		}

		if err := json.Unmarshal(data, &session); err != nil {
			fmt.Printf("  ⚠️ Failed to parse %s: %v\n", file.Name(), err)
			continue
		}

		// Save all messages to DB
		for _, m := range session.Messages {
			if err := sm.db.SaveMessage(session.Key, m.Role, m.Content); err != nil {
				fmt.Printf("  ❌ Failed to save message from %s: %v\n", session.Key, err)
			}
		}

		// Success - Rename file to .bak or delete
		os.Rename(path, path+".bak")
		count++
	}

	if count > 0 {
		fmt.Printf("✅ Migration complete. %d sessions moved to history.db\n", count)
	}
	return nil
}
