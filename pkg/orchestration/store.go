package orchestration

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "modernc.org/sqlite"
)

var ErrConflict = errors.New("다른 변경이 먼저 저장되었습니다. 새로고침 후 다시 시도하세요")
var ErrNotFound = errors.New("항목을 찾을 수 없습니다")

type Store struct {
	db    *sql.DB
	Owner string
}

func OpenStore(root, owner string) (*Store, error) {
	if strings.TrimSpace(owner) == "" {
		return nil, errors.New("사용자 식별자가 필요합니다")
	}
	hash := sha256.Sum256([]byte(owner))
	dir := filepath.Join(root, "users", hex.EncodeToString(hash[:]))
	if err := os.MkdirAll(dir, 0700); err != nil {
		return nil, err
	}
	db, err := sql.Open("sqlite", filepath.Join(dir, "state.sqlite"))
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	s := &Store{db: db, Owner: owner}
	_, err = db.Exec(`PRAGMA journal_mode=WAL; PRAGMA busy_timeout=5000;
 CREATE TABLE IF NOT EXISTS records(kind TEXT NOT NULL,id TEXT NOT NULL,data TEXT NOT NULL,PRIMARY KEY(kind,id));
 CREATE TABLE IF NOT EXISTS events(id INTEGER PRIMARY KEY AUTOINCREMENT,run_id TEXT,task_id TEXT,type TEXT,message TEXT,created TEXT);
 CREATE TABLE IF NOT EXISTS memory(id TEXT PRIMARY KEY,revision INTEGER NOT NULL,data TEXT NOT NULL);
 CREATE TABLE IF NOT EXISTS memory_history(id TEXT,revision INTEGER,data TEXT,PRIMARY KEY(id,revision));`)
	if err != nil {
		db.Close()
		return nil, err
	}
	return s, nil
}
func (s *Store) Close() error { return s.db.Close() }
func (s *Store) Put(kind, id string, v interface{}) error {
	b, err := json.Marshal(v)
	if err != nil {
		return err
	}
	_, err = s.db.Exec(`INSERT INTO records(kind,id,data) VALUES(?,?,?) ON CONFLICT(kind,id) DO UPDATE SET data=excluded.data`, kind, id, b)
	return err
}
func (s *Store) Get(kind, id string, v interface{}) error {
	var b []byte
	err := s.db.QueryRow(`SELECT data FROM records WHERE kind=? AND id=?`, kind, id).Scan(&b)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrNotFound
	}
	if err != nil {
		return err
	}
	return json.Unmarshal(b, v)
}
func (s *Store) List(kind string, v interface{}) error {
	rows, err := s.db.Query(`SELECT data FROM records WHERE kind=? ORDER BY rowid DESC`, kind)
	if err != nil {
		return err
	}
	defer rows.Close()
	items := []json.RawMessage{}
	for rows.Next() {
		var b []byte
		if err = rows.Scan(&b); err != nil {
			return err
		}
		items = append(items, json.RawMessage(b))
	}
	if err = rows.Err(); err != nil {
		return err
	}
	b, _ := json.Marshal(items)
	return json.Unmarshal(b, v)
}
func (s *Store) Event(run, task, kind, message string) error {
	_, err := s.db.Exec(`INSERT INTO events(run_id,task_id,type,message,created) VALUES(?,?,?,?,?)`, run, task, kind, message, time.Now().UTC().Format(time.RFC3339Nano))
	return err
}
func (s *Store) Events(after int64, run string) ([]Event, error) {
	rows, err := s.db.Query(`SELECT id,run_id,task_id,type,message,created FROM events WHERE id>? AND (?='' OR run_id=?) ORDER BY id LIMIT 500`, after, run, run)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []Event{}
	for rows.Next() {
		var e Event
		var ts string
		if err = rows.Scan(&e.ID, &e.RunID, &e.TaskID, &e.Type, &e.Message, &ts); err != nil {
			return nil, err
		}
		e.Time, _ = time.Parse(time.RFC3339Nano, ts)
		items = append(items, e)
	}
	return items, rows.Err()
}

func (s *Store) SaveMemory(m Memory) (Memory, error) {
	now := time.Now().UTC()
	if len(m.Title) == 0 || len(m.Title) > 300 || len(m.Content) == 0 || len(m.Content) > 65536 {
		return m, errors.New("제목과 본문을 입력하세요 (제목 300바이트, 본문 64KiB 이내)")
	}
	if !Contains([]string{"short", "long", "wiki"}, m.Kind) || !Contains([]string{"private", "shared"}, m.Visibility) {
		return m, errors.New("기억 종류와 공유 범위를 확인하세요")
	}
	if m.Visibility == "private" && m.NodeID == "" && m.SessionID == "" {
		return m, errors.New("전용 기억에는 장비 또는 세션이 필요합니다")
	}
	if m.Kind == "wiki" && !m.Verified {
		return m, errors.New("Wiki는 검증된 내용만 저장할 수 있습니다")
	}
	if m.Kind == "short" && m.ExpiresAt == nil {
		expiry := now.Add(24 * time.Hour)
		m.ExpiresAt = &expiry
	}
	if m.ExpiresAt != nil && !m.ExpiresAt.After(now) {
		return m, errors.New("만료 시간은 현재보다 이후여야 합니다")
	}
	if m.ID == "" {
		m.ID = NewID()
		m.Revision = 0
		m.CreatedAt = now
	}
	tx, err := s.db.Begin()
	if err != nil {
		return m, err
	}
	defer tx.Rollback()
	if m.Revision > 0 {
		var old []byte
		if err = tx.QueryRow(`SELECT data FROM memory WHERE id=? AND revision=?`, m.ID, m.Revision).Scan(&old); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return m, ErrConflict
			}
			return m, err
		}
		var prev Memory
		json.Unmarshal(old, &prev)
		m.CreatedAt = prev.CreatedAt
	}
	expected := m.Revision
	m.Revision++
	m.UpdatedAt = now
	b, _ := json.Marshal(m)
	if expected == 0 {
		_, err = tx.Exec(`INSERT INTO memory(id,revision,data) VALUES(?,?,?)`, m.ID, m.Revision, b)
	} else {
		var res sql.Result
		res, err = tx.Exec(`UPDATE memory SET revision=?,data=? WHERE id=? AND revision=?`, m.Revision, b, m.ID, expected)
		if err == nil {
			n, _ := res.RowsAffected()
			if n != 1 {
				return m, ErrConflict
			}
		}
	}
	if err != nil {
		return m, err
	}
	if _, err = tx.Exec(`INSERT INTO memory_history(id,revision,data) VALUES(?,?,?)`, m.ID, m.Revision, b); err != nil {
		return m, err
	}
	return m, tx.Commit()
}
func (s *Store) Memories(query, node, session string, worker bool) ([]Memory, error) {
	rows, err := s.db.Query(`SELECT data FROM memory ORDER BY rowid DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Memory{}
	q := strings.ToLower(query)
	now := time.Now()
	for rows.Next() {
		var b []byte
		var m Memory
		if err = rows.Scan(&b); err != nil {
			return nil, err
		}
		if err = json.Unmarshal(b, &m); err != nil {
			return nil, err
		}
		if m.ExpiresAt != nil && !m.ExpiresAt.After(now) {
			continue
		}
		if worker && m.Visibility == "private" {
			if m.NodeID == "" || m.NodeID != node || m.SessionID != "" && m.SessionID != session {
				continue
			}
		}
		if q != "" && !strings.Contains(strings.ToLower(m.Title+" "+m.Content), q) {
			continue
		}
		out = append(out, m)
		if len(out) >= 200 {
			break
		}
	}
	return out, rows.Err()
}
func (s *Store) DeleteMemory(id string, revision int) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	res, err := tx.Exec(`DELETE FROM memory WHERE id=? AND revision=?`, id, revision)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrConflict
	}
	if _, err = tx.Exec(`DELETE FROM memory_history WHERE id=?`, id); err != nil {
		return err
	}
	return tx.Commit()
}
func (s *Store) PruneMemories() error {
	items, err := s.db.Query(`SELECT id,data FROM memory`)
	if err != nil {
		return err
	}
	ids := []string{}
	for items.Next() {
		var id string
		var b []byte
		if err = items.Scan(&id, &b); err != nil {
			items.Close()
			return err
		}
		var m Memory
		json.Unmarshal(b, &m)
		if m.ExpiresAt != nil && !m.ExpiresAt.After(time.Now()) {
			ids = append(ids, id)
		}
	}
	err = items.Err()
	items.Close()
	if err != nil {
		return err
	}
	for _, id := range ids {
		if _, err = s.db.Exec(`DELETE FROM memory WHERE id=?`, id); err != nil {
			return err
		}
		if _, err = s.db.Exec(`DELETE FROM memory_history WHERE id=?`, id); err != nil {
			return err
		}
	}
	return nil
}
func (s *Store) MemoryHistory(id string) ([]Memory, error) {
	rows, err := s.db.Query(`SELECT data FROM memory_history WHERE id=? ORDER BY revision DESC`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Memory{}
	for rows.Next() {
		var b []byte
		var m Memory
		if err = rows.Scan(&b); err != nil {
			return nil, err
		}
		if err = json.Unmarshal(b, &m); err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, rows.Err()
}
func (s *Store) Promote(id string, revision int) (Memory, error) {
	var b []byte
	if err := s.db.QueryRow(`SELECT data FROM memory WHERE id=? AND revision=?`, id, revision).Scan(&b); err != nil {
		return Memory{}, ErrConflict
	}
	var m Memory
	if err := json.Unmarshal(b, &m); err != nil {
		return m, err
	}
	if m.ExpiresAt != nil && !m.ExpiresAt.After(time.Now()) {
		return m, ErrNotFound
	}
	m.Kind = "wiki"
	m.Verified = true
	m.ExpiresAt = nil
	m.Source = fmt.Sprintf("기억 %s revision %d 검토 후 승격; %s", id, revision, m.Source)
	return s.SaveMemory(m)
}
