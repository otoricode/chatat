// Local message repository — CRUD for messages table
import { getDatabase } from './index';
import type { LocalMessage } from './types';

function generateId(): string {
  return `local_${Date.now()}_${Math.random().toString(36).slice(2, 9)}`;
}

/**
 * Insert a new message.
 * Returns the local id.
 */
export async function insertMessage(msg: Omit<LocalMessage, 'id'>): Promise<string> {
  const db = await getDatabase();
  const id = generateId();

  await db.runAsync(
    `INSERT INTO messages (id, server_id, chat_id, sender_id, sender_name,
      content, type, status, reply_to_id, metadata, is_deleted, is_pending, created_at)
    VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
    [
      id, msg.server_id, msg.chat_id, msg.sender_id, msg.sender_name,
      msg.content, msg.type, msg.status, msg.reply_to_id, msg.metadata,
      msg.is_deleted, msg.is_pending, msg.created_at,
    ]
  );
  return id;
}

/**
 * Upsert a message by server_id.
 * Used during sync — inserts if not exists, updates if exists.
 */
export async function upsertMessage(msg: Omit<LocalMessage, 'id'>): Promise<string> {
  if (!msg.server_id) {
    return insertMessage(msg);
  }

  const db = await getDatabase();
  const existing = await db.getFirstAsync<LocalMessage>(
    'SELECT id FROM messages WHERE server_id = ?',
    [msg.server_id]
  );

  if (existing) {
    await db.runAsync(
      `UPDATE messages SET
        content = ?, type = ?, status = ?, is_deleted = ?,
        is_pending = ?, metadata = ?
      WHERE server_id = ?`,
      [
        msg.content, msg.type, msg.status, msg.is_deleted,
        msg.is_pending, msg.metadata, msg.server_id,
      ]
    );
    return existing.id;
  }

  return insertMessage(msg);
}

/**
 * Get messages for a chat, paginated.
 */
export async function getMessages(
  chatId: string,
  limit = 50,
  beforeTimestamp?: number
): Promise<LocalMessage[]> {
  const db = await getDatabase();

  if (beforeTimestamp) {
    return db.getAllAsync<LocalMessage>(
      'SELECT * FROM messages WHERE chat_id = ? AND created_at < ? ORDER BY created_at DESC LIMIT ?',
      [chatId, beforeTimestamp, limit]
    );
  }

  return db.getAllAsync<LocalMessage>(
    'SELECT * FROM messages WHERE chat_id = ? ORDER BY created_at DESC LIMIT ?',
    [chatId, limit]
  );
}

/**
 * Get pending (unsent) messages.
 */
export async function getPendingMessages(): Promise<LocalMessage[]> {
  const db = await getDatabase();
  return db.getAllAsync<LocalMessage>(
    'SELECT * FROM messages WHERE is_pending = 1 ORDER BY created_at ASC'
  );
}

/**
 * Get pending messages for a specific chat.
 */
export async function getPendingMessagesForChat(chatId: string): Promise<LocalMessage[]> {
  const db = await getDatabase();
  return db.getAllAsync<LocalMessage>(
    'SELECT * FROM messages WHERE chat_id = ? AND is_pending = 1 ORDER BY created_at ASC',
    [chatId]
  );
}

/**
 * Mark a message as sent (update server_id and status).
 */
export async function markMessageSent(localId: string, serverId: string): Promise<void> {
  const db = await getDatabase();
  await db.runAsync(
    'UPDATE messages SET server_id = ?, status = ?, is_pending = 0 WHERE id = ?',
    [serverId, 'sent', localId]
  );
}

/**
 * Mark a message as failed.
 */
export async function markMessageFailed(localId: string): Promise<void> {
  const db = await getDatabase();
  await db.runAsync(
    "UPDATE messages SET status = 'failed' WHERE id = ?",
    [localId]
  );
}

/**
 * Update message status (delivered, read).
 */
export async function updateMessageStatus(serverId: string, status: string): Promise<void> {
  const db = await getDatabase();
  await db.runAsync(
    'UPDATE messages SET status = ? WHERE server_id = ?',
    [status, serverId]
  );
}

/**
 * Soft-delete a message.
 */
export async function softDeleteMessage(serverId: string): Promise<void> {
  const db = await getDatabase();
  await db.runAsync(
    'UPDATE messages SET is_deleted = 1, content = NULL WHERE server_id = ?',
    [serverId]
  );
}

/**
 * Get the latest message server_id for a chat (for sync cursor).
 */
export async function getLatestMessageServerId(chatId: string): Promise<string | null> {
  const db = await getDatabase();
  const result = await db.getFirstAsync<{ server_id: string }>(
    'SELECT server_id FROM messages WHERE chat_id = ? AND server_id IS NOT NULL ORDER BY created_at DESC LIMIT 1',
    [chatId]
  );
  return result?.server_id ?? null;
}

/**
 * Get message count for a chat.
 */
export async function getMessageCount(chatId: string): Promise<number> {
  const db = await getDatabase();
  const result = await db.getFirstAsync<{ count: number }>(
    'SELECT COUNT(*) as count FROM messages WHERE chat_id = ?',
    [chatId]
  );
  return result?.count ?? 0;
}

/**
 * Delete all messages for a chat.
 */
export async function deleteMessagesForChat(chatId: string): Promise<void> {
  const db = await getDatabase();
  await db.runAsync('DELETE FROM messages WHERE chat_id = ?', [chatId]);
}
