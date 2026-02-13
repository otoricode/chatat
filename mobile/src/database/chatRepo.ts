// Local chat repository — CRUD for chats table
import { getDatabase } from './index';
import type { LocalChat } from './types';

function generateId(): string {
  return `local_${Date.now()}_${Math.random().toString(36).slice(2, 9)}`;
}

/**
 * Upsert a chat by server_id.
 * Inserts if not exists, updates if exists.
 */
export async function upsertChat(chat: Omit<LocalChat, 'id'>): Promise<string> {
  const db = await getDatabase();

  const existing = await db.getFirstAsync<LocalChat>(
    'SELECT id FROM chats WHERE server_id = ?',
    [chat.server_id]
  );

  if (existing) {
    await db.runAsync(
      `UPDATE chats SET
        type = ?, name = ?, icon = ?, last_message = ?,
        last_message_at = ?, unread_count = ?, is_muted = ?,
        is_archived = ?, pinned_at = ?, synced_at = ?, created_at = ?
      WHERE server_id = ?`,
      [
        chat.type, chat.name, chat.icon, chat.last_message,
        chat.last_message_at, chat.unread_count, chat.is_muted,
        chat.is_archived, chat.pinned_at, chat.synced_at, chat.created_at,
        chat.server_id,
      ]
    );
    return existing.id;
  }

  const id = generateId();
  await db.runAsync(
    `INSERT INTO chats (id, server_id, type, name, icon, last_message,
      last_message_at, unread_count, is_muted, is_archived, pinned_at,
      synced_at, created_at)
    VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
    [
      id, chat.server_id, chat.type, chat.name, chat.icon, chat.last_message,
      chat.last_message_at, chat.unread_count, chat.is_muted,
      chat.is_archived, chat.pinned_at, chat.synced_at, chat.created_at,
    ]
  );
  return id;
}

/**
 * Get all chats ordered by most recent message.
 */
export async function getChats(): Promise<LocalChat[]> {
  const db = await getDatabase();
  return db.getAllAsync<LocalChat>(
    'SELECT * FROM chats ORDER BY COALESCE(pinned_at, 0) DESC, COALESCE(last_message_at, created_at) DESC'
  );
}

/**
 * Get a chat by server_id.
 */
export async function getChatByServerId(serverId: string): Promise<LocalChat | null> {
  const db = await getDatabase();
  return db.getFirstAsync<LocalChat>(
    'SELECT * FROM chats WHERE server_id = ?',
    [serverId]
  );
}

/**
 * Update unread count for a chat.
 */
export async function updateUnreadCount(serverId: string, count: number): Promise<void> {
  const db = await getDatabase();
  await db.runAsync(
    'UPDATE chats SET unread_count = ? WHERE server_id = ?',
    [count, serverId]
  );
}

/**
 * Delete a chat and its messages.
 */
export async function deleteChat(serverId: string): Promise<void> {
  const db = await getDatabase();
  await db.runAsync('DELETE FROM messages WHERE chat_id = ?', [serverId]);
  await db.runAsync('DELETE FROM chats WHERE server_id = ?', [serverId]);
}

/**
 * Get count of all chats.
 */
export async function getChatCount(): Promise<number> {
  const db = await getDatabase();
  const result = await db.getFirstAsync<{ count: number }>(
    'SELECT COUNT(*) as count FROM chats'
  );
  return result?.count ?? 0;
}
