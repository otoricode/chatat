// Local database schema definitions
// Tables mirror server models for offline-first support

export const DB_VERSION = 1;

export const CREATE_TABLES_SQL = `
  CREATE TABLE IF NOT EXISTS chats (
    id TEXT PRIMARY KEY,
    server_id TEXT UNIQUE NOT NULL,
    type TEXT NOT NULL CHECK(type IN ('personal', 'group')),
    name TEXT,
    icon TEXT,
    last_message TEXT,
    last_message_at INTEGER,
    unread_count INTEGER NOT NULL DEFAULT 0,
    is_muted INTEGER NOT NULL DEFAULT 0,
    is_archived INTEGER NOT NULL DEFAULT 0,
    pinned_at INTEGER,
    synced_at INTEGER NOT NULL DEFAULT 0,
    created_at INTEGER NOT NULL DEFAULT 0
  );

  CREATE INDEX IF NOT EXISTS idx_chats_server_id ON chats(server_id);
  CREATE INDEX IF NOT EXISTS idx_chats_last_message_at ON chats(last_message_at DESC);

  CREATE TABLE IF NOT EXISTS messages (
    id TEXT PRIMARY KEY,
    server_id TEXT UNIQUE,
    chat_id TEXT NOT NULL,
    sender_id TEXT NOT NULL,
    sender_name TEXT NOT NULL DEFAULT '',
    content TEXT,
    type TEXT NOT NULL DEFAULT 'text',
    status TEXT NOT NULL DEFAULT 'sending',
    reply_to_id TEXT,
    metadata TEXT,
    is_deleted INTEGER NOT NULL DEFAULT 0,
    is_pending INTEGER NOT NULL DEFAULT 0,
    created_at INTEGER NOT NULL DEFAULT 0
  );

  CREATE INDEX IF NOT EXISTS idx_messages_server_id ON messages(server_id);
  CREATE INDEX IF NOT EXISTS idx_messages_chat_id ON messages(chat_id, created_at DESC);
  CREATE INDEX IF NOT EXISTS idx_messages_pending ON messages(is_pending) WHERE is_pending = 1;

  CREATE TABLE IF NOT EXISTS contacts (
    id TEXT PRIMARY KEY,
    server_id TEXT UNIQUE NOT NULL,
    user_id TEXT NOT NULL,
    name TEXT NOT NULL,
    phone TEXT NOT NULL,
    avatar TEXT,
    status_text TEXT,
    is_registered INTEGER NOT NULL DEFAULT 0,
    synced_at INTEGER NOT NULL DEFAULT 0
  );

  CREATE INDEX IF NOT EXISTS idx_contacts_server_id ON contacts(server_id);
  CREATE INDEX IF NOT EXISTS idx_contacts_name ON contacts(name);

  CREATE TABLE IF NOT EXISTS documents (
    id TEXT PRIMARY KEY,
    server_id TEXT UNIQUE NOT NULL,
    title TEXT NOT NULL DEFAULT '',
    icon TEXT,
    locked INTEGER NOT NULL DEFAULT 0,
    lock_type TEXT,
    owner_id TEXT NOT NULL,
    owner_name TEXT NOT NULL DEFAULT '',
    context_type TEXT NOT NULL DEFAULT 'standalone',
    context_id TEXT,
    updated_at INTEGER NOT NULL DEFAULT 0,
    synced_at INTEGER NOT NULL DEFAULT 0
  );

  CREATE INDEX IF NOT EXISTS idx_documents_server_id ON documents(server_id);
  CREATE INDEX IF NOT EXISTS idx_documents_updated_at ON documents(updated_at DESC);

  CREATE TABLE IF NOT EXISTS sync_meta (
    key TEXT PRIMARY KEY,
    value TEXT NOT NULL,
    updated_at INTEGER NOT NULL DEFAULT 0
  );

  CREATE TABLE IF NOT EXISTS pending_operations (
    id TEXT PRIMARY KEY,
    type TEXT NOT NULL,
    entity_type TEXT NOT NULL,
    entity_id TEXT NOT NULL,
    payload TEXT NOT NULL,
    retry_count INTEGER NOT NULL DEFAULT 0,
    status TEXT NOT NULL DEFAULT 'pending',
    created_at INTEGER NOT NULL DEFAULT 0,
    last_attempt_at INTEGER
  );

  CREATE INDEX IF NOT EXISTS idx_pending_ops_status ON pending_operations(status);
`;

export const DROP_TABLES_SQL = `
  DROP TABLE IF EXISTS pending_operations;
  DROP TABLE IF EXISTS sync_meta;
  DROP TABLE IF EXISTS documents;
  DROP TABLE IF EXISTS contacts;
  DROP TABLE IF EXISTS messages;
  DROP TABLE IF EXISTS chats;
`;
