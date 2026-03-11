package session

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/dirmich/marubot/pkg/providers"
	_ "modernc.org/sqlite"
)

type SQLiteStore struct {
	db *sql.DB
}

func NewSQLiteStore(dbPath string) (*SQLiteStore, error) {
	dir := filepath.Dir(dbPath)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		os.MkdirAll(dir, 0755)
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, err
	}

	store := &SQLiteStore{db: db}
	if err := store.init(); err != nil {
		db.Close()
		return nil, err
	}

	return store, nil
}

func (s *SQLiteStore) init() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS sessions (
			key TEXT PRIMARY KEY,
			created TIMESTAMP,
			updated TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS messages (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			session_key TEXT,
			role TEXT,
			content TEXT,
			tokens INTEGER DEFAULT 0,
			created_at TIMESTAMP,
			FOREIGN KEY(session_key) REFERENCES sessions(key) ON DELETE CASCADE
		)`,
		`CREATE INDEX IF NOT EXISTS idx_messages_session ON messages(session_key)`,
		
		// 🧠 Facts (Long-term Memory Directives/Preferences)
		`CREATE TABLE IF NOT EXISTS facts (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			category TEXT, -- preference, rule, project_fact, user_info
			content TEXT,
			confidence REAL,
			source_message_id INTEGER,
			status TEXT DEFAULT 'active', -- active, superseded, archived
			created_at TIMESTAMP,
			updated_at TIMESTAMP,
			expires_at TIMESTAMP
		)`,

		// 📚 Memory Chunks (For Contextual Retrieval)
		`CREATE TABLE IF NOT EXISTS memory_chunks (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			session_key TEXT,
			start_msg_id INTEGER,
			end_msg_id INTEGER,
			content TEXT, -- Aggregated content of the window
			summary TEXT, -- Optional summary of the window
			created_at TIMESTAMP
		)`,

		// FTS5 table for RAG (Searching across messages and chunks)
		`CREATE VIRTUAL TABLE IF NOT EXISTS memory_fts USING fts5(
			content,
			source_id UNINDEXED, -- references messages(id) or memory_chunks(id)
			source_type UNINDEXED -- 'message' or 'chunk'
		)`,

		// Triggers to keep FTS index in sync for messages
		`CREATE TRIGGER IF NOT EXISTS messages_ai AFTER INSERT ON messages BEGIN
			INSERT INTO memory_fts(content, source_id, source_type) VALUES (new.content, new.id, 'message');
		END`,
	}

	for _, q := range queries {
		if _, err := s.db.Exec(q); err != nil {
			return err
		}
	}
	return nil
}

func (s *SQLiteStore) Close() error {
	return s.db.Close()
}

func (s *SQLiteStore) SaveMessage(sessionKey, role, content string) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Ensure session exists
	now := time.Now()
	_, err = tx.Exec(`
		INSERT INTO sessions (key, created, updated) 
		VALUES (?, ?, ?) 
		ON CONFLICT(key) DO UPDATE SET updated = ?`,
		sessionKey, now, now, now)
	if err != nil {
		return err
	}

	// Insert message
	_, err = tx.Exec(`
		INSERT INTO messages (session_key, role, content, created_at) 
		VALUES (?, ?, ?, ?)`,
		sessionKey, role, content, now)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (s *SQLiteStore) GetMessages(sessionKey string, limit int) ([]providers.Message, error) {
	query := `SELECT role, content FROM messages WHERE session_key = ? ORDER BY id DESC`
	if limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", limit)
	}

	rows, err := s.db.Query(query, sessionKey)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var msgs []providers.Message
	for rows.Next() {
		var m providers.Message
		if err := rows.Scan(&m.Role, &m.Content); err != nil {
			return nil, err
		}
		msgs = append([]providers.Message{m}, msgs...)
	}
	return msgs, nil
}

func (s *SQLiteStore) SearchRelevant(query string, limit int) ([]providers.Message, error) {
	// 📚 Combined search across messages and memory chunks
	sqlQuery := `
		SELECT f.source_type, m.role, m.content 
		FROM memory_fts f
		LEFT JOIN messages m ON f.source_id = m.id AND f.source_type = 'message'
		WHERE memory_fts MATCH ? 
		ORDER BY rank 
		LIMIT ?`
	
	rows, err := s.db.Query(sqlQuery, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var msgs []providers.Message
	for rows.Next() {
		var srcType string
		var m providers.Message
		if err := rows.Scan(&srcType, &m.Role, &m.Content); err != nil {
			return nil, err
		}
		
		// If it's a chunk, it might not have a single 'Role'
		if srcType == "chunk" {
			m.Role = "assistant" // Default for summarized context
		}
		
		msgs = append(msgs, m)
	}
	return msgs, nil
}

// 🧘 GetActiveFacts retrieves the most relevant rules, preferences, and facts
func (s *SQLiteStore) GetActiveFacts(category string) ([]string, error) {
	query := `SELECT content FROM facts WHERE status = 'active'`
	var args []interface{}
	if category != "" {
		query += " AND category = ?"
		args = append(args, category)
	}
	query += " ORDER BY confidence DESC, updated_at DESC LIMIT 10"

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var facts []string
	for rows.Next() {
		var f string
		if err := rows.Scan(&f); err != nil {
			return nil, err
		}
		facts = append(facts, f)
	}
	return facts, nil
}

func (s *SQLiteStore) SaveFact(category, content string, confidence float64, srcID int) error {
	now := time.Now()
	_, err := s.db.Exec(`
		INSERT INTO facts (category, content, confidence, source_message_id, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?)`,
		category, content, confidence, srcID, now, now)
	return err
}

func (s *SQLiteStore) GetAllSessions() ([]string, error) {
	rows, err := s.db.Query(`SELECT key FROM sessions ORDER BY updated DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var keys []string
	for rows.Next() {
		var k string
		if err := rows.Scan(&k); err != nil {
			return nil, err
		}
		keys = append(keys, k)
	}
	return keys, nil
}
