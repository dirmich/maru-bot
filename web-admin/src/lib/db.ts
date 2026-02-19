import Database from 'better-sqlite3';
import { drizzle } from 'drizzle-orm/better-sqlite3';
import path from 'path';
import os from 'os';
import fs from 'fs';
import * as schema from './schema';
import { desc, eq } from 'drizzle-orm';

const DB_DIR = path.join(os.homedir(), '.marubot', 'web');
const DB_PATH = path.join(DB_DIR, 'admin.db');

export let db: any = null;

function initDb() {
  if (db) return db;

  // Build-time safety
  if (typeof window !== 'undefined' || process.env.NEXT_PHASE === 'phase-production-build') {
    return null;
  }

  if (!fs.existsSync(DB_DIR)) {
    fs.mkdirSync(DB_DIR, { recursive: true });
  }

  try {
    const sqlite = new Database(DB_PATH);
    db = drizzle(sqlite, { schema });

    // Auto-create tables if they don't exist (Simple way for now)
    sqlite.exec(`
      CREATE TABLE IF NOT EXISTS user (
        id TEXT PRIMARY KEY NOT NULL,
        name TEXT,
        email TEXT NOT NULL,
        emailVerified INTEGER,
        image TEXT
      );
      CREATE TABLE IF NOT EXISTS account (
        userId TEXT NOT NULL REFERENCES user(id) ON DELETE CASCADE,
        type TEXT NOT NULL,
        provider TEXT NOT NULL,
        providerAccountId TEXT NOT NULL,
        refresh_token TEXT,
        access_token TEXT,
        expires_at INTEGER,
        token_type TEXT,
        scope TEXT,
        id_token TEXT,
        session_state TEXT,
        PRIMARY KEY (provider, providerAccountId)
      );
      CREATE TABLE IF NOT EXISTS session (
        sessionToken TEXT PRIMARY KEY NOT NULL,
        userId TEXT NOT NULL REFERENCES user(id) ON DELETE CASCADE,
        expires INTEGER NOT NULL
      );
      CREATE TABLE IF NOT EXISTS verificationToken (
        identifier TEXT NOT NULL,
        token TEXT NOT NULL,
        expires INTEGER NOT NULL,
        PRIMARY KEY (identifier, token)
      );
      CREATE TABLE IF NOT EXISTS settings (
        key TEXT PRIMARY KEY NOT NULL,
        value TEXT NOT NULL
      );
      CREATE TABLE IF NOT EXISTS messages (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        role TEXT NOT NULL,
        content TEXT NOT NULL,
        timestamp INTEGER DEFAULT (strftime('%s', 'now') * 1000)
      );
    `);

    return db;
  } catch (e) {
    console.warn('DB initialization deferred or failed:', e);
    return null;
  }
}

export const getDb = initDb;

export interface Message {
  id: number;
  role: 'user' | 'assistant' | 'system';
  content: string;
  timestamp: Date | null;
}

export async function getMessages(limit = 50): Promise<Message[]> {
  const conn = initDb();
  if (!conn) return [];
  const results = conn.select().from(schema.messages).orderBy(desc(schema.messages.timestamp)).limit(limit).all();
  return results as Message[];
}

export async function addMessage(role: string, content: string) {
  const conn = initDb();
  if (!conn) return;
  return conn.insert(schema.messages).values({ role, content }).run();
}

export async function clearMessages() {
  const conn = initDb();
  if (!conn) return;
  return conn.delete(schema.messages).run();
}

export async function getSetting(key: string): Promise<string | null> {
  const conn = initDb();
  if (!conn) return null;
  const result = conn.select().from(schema.settings).where(eq(schema.settings.key, key)).get();
  return result ? result.value : null;
}

export async function setSetting(key: string, value: string) {
  const conn = initDb();
  if (!conn) return;
  return conn.insert(schema.settings).values({ key, value }).onConflictDoUpdate({
    target: schema.settings.key,
    set: { value }
  }).run();
}
