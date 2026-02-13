import Database from 'better-sqlite3';
import path from 'path';
import os from 'os';
import fs from 'fs';

const DB_DIR = path.join(os.homedir(), '.marubot', 'web');
const DB_PATH = path.join(DB_DIR, 'admin.db');

let _db: any = null;

function getDb() {
  if (_db) return _db;

  // Build-time safety
  if (typeof window !== 'undefined' || process.env.NEXT_PHASE === 'phase-production-build') {
    return null;
  }

  if (!fs.existsSync(DB_DIR)) {
    fs.mkdirSync(DB_DIR, { recursive: true });
  }

  try {
    _db = new Database(DB_PATH);
    // Initialize tables
    _db.exec(`
          CREATE TABLE IF NOT EXISTS messages (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            role TEXT NOT NULL,
            content TEXT NOT NULL,
            timestamp DATETIME DEFAULT CURRENT_TIMESTAMP
          );
          
          CREATE TABLE IF NOT EXISTS settings (
            key TEXT PRIMARY KEY,
            value TEXT NOT NULL
          );
        `);
    return _db;
  } catch (e) {
    console.warn('DB initialization deferred or failed:', e);
    return null;
  }
}

export default getDb(); // Fallback for existing imports, but better use getDb()

export interface Message {
  id: number;
  role: 'user' | 'assistant' | 'system';
  content: string;
  timestamp: string;
}

export function getMessages(limit = 50): Message[] {
  const db = getDb();
  if (!db) return [];
  return db.prepare('SELECT * FROM messages ORDER BY timestamp DESC LIMIT ?').all(limit) as Message[];
}

export function addMessage(role: string, content: string) {
  const db = getDb();
  if (!db) return;
  return db.prepare('INSERT INTO messages (role, content) VALUES (?, ?)').run(role, content);
}

export function clearMessages() {
  const db = getDb();
  if (!db) return;
  return db.prepare('DELETE FROM messages').run();
}

export function getSetting(key: string): string | null {
  const db = getDb();
  if (!db) return null;
  const row = db.prepare('SELECT value FROM settings WHERE key = ?').get(key) as { value: string } | undefined;
  return row ? row.value : null;
}

export function setSetting(key: string, value: string) {
  const db = getDb();
  if (!db) return;
  return db.prepare('INSERT OR REPLACE INTO settings (key, value) VALUES (?, ?)').run(key, value);
}
