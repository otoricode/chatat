// Local contact repository — CRUD for contacts table
import { getDatabase } from './index';
import type { LocalContact } from './types';

function generateId(): string {
  return `local_${Date.now()}_${Math.random().toString(36).slice(2, 9)}`;
}

/**
 * Upsert a contact by server_id.
 */
export async function upsertContact(contact: Omit<LocalContact, 'id'>): Promise<string> {
  const db = await getDatabase();

  const existing = await db.getFirstAsync<LocalContact>(
    'SELECT id FROM contacts WHERE server_id = ?',
    [contact.server_id]
  );

  if (existing) {
    await db.runAsync(
      `UPDATE contacts SET
        user_id = ?, name = ?, phone = ?, avatar = ?,
        status_text = ?, is_registered = ?, synced_at = ?
      WHERE server_id = ?`,
      [
        contact.user_id, contact.name, contact.phone, contact.avatar,
        contact.status_text, contact.is_registered, contact.synced_at,
        contact.server_id,
      ]
    );
    return existing.id;
  }

  const id = generateId();
  await db.runAsync(
    `INSERT INTO contacts (id, server_id, user_id, name, phone, avatar,
      status_text, is_registered, synced_at)
    VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
    [
      id, contact.server_id, contact.user_id, contact.name, contact.phone,
      contact.avatar, contact.status_text, contact.is_registered, contact.synced_at,
    ]
  );
  return id;
}

/**
 * Get all contacts ordered by name.
 */
export async function getContacts(): Promise<LocalContact[]> {
  const db = await getDatabase();
  return db.getAllAsync<LocalContact>(
    'SELECT * FROM contacts ORDER BY name ASC'
  );
}

/**
 * Get registered contacts only.
 */
export async function getRegisteredContacts(): Promise<LocalContact[]> {
  const db = await getDatabase();
  return db.getAllAsync<LocalContact>(
    'SELECT * FROM contacts WHERE is_registered = 1 ORDER BY name ASC'
  );
}

/**
 * Search contacts by name.
 */
export async function searchContacts(query: string): Promise<LocalContact[]> {
  const db = await getDatabase();
  return db.getAllAsync<LocalContact>(
    "SELECT * FROM contacts WHERE name LIKE ? ORDER BY name ASC LIMIT 20",
    [`%${query}%`]
  );
}

/**
 * Delete a contact.
 */
export async function deleteContact(serverId: string): Promise<void> {
  const db = await getDatabase();
  await db.runAsync('DELETE FROM contacts WHERE server_id = ?', [serverId]);
}

/**
 * Get contact count.
 */
export async function getContactCount(): Promise<number> {
  const db = await getDatabase();
  const result = await db.getFirstAsync<{ count: number }>(
    'SELECT COUNT(*) as count FROM contacts'
  );
  return result?.count ?? 0;
}
