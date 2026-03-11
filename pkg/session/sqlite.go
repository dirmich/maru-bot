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
			created_at TIMESTAMP,
			FOREIGN KEY(session_key) REFERENCES sessions(key) ON DELETE CASCADE
		)`,
		`CREATE INDEX IF NOT EXISTS idx_messages_session ON messages(session_key)`,
		// FTS5 table for RAG
		`CREATE VIRTUAL TABLE IF NOT EXISTS messages_fts USING fts5(
			content,
			content='messages',
			content_rowid='id'
		)`,
		// Triggers to keep FTS index in sync
		`CREATE TRIGGER IF NOT EXISTS messages_ai AFTER INSERT ON messages BEGIN
			INSERT INTO messages_fts(rowid, content) VALUES (new.id, new.content);
		END`,
		`CREATE TRIGGER IF NOT EXISTS messages_ad AFTER DELETE ON messages BEGIN
			INSERT INTO messages_fts(messages_fts, rowid, content) VALUES('delete', old.id, old.content);
		END`,
		`CREATE TRIGGER IF NOT EXISTS messages_au AFTER UPDATE ON messages BEGIN
			INSERT INTO messages_fts(messages_fts, rowid, content) VALUES('delete', old.id, old.content);
			INSERT INTO messages_fts(rowid, content) VALUES (new.id, new.content);
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
	// Simple BM25 search via FTS5
	sqlQuery := `
		SELECT m.role, m.content 
		FROM messages m
		JOIN messages_fts f ON m.id = f.rowid
		WHERE messages_fts MATCH ? 
		ORDER BY rank 
		LIMIT ?`
	
	rows, err := s.db.Query(sqlQuery, query, limit)
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
		msgs = append(msgs, m)
	}
	return msgs, nil
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
