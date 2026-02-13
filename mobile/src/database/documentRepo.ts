// Local document repository — CRUD for documents table
import { getDatabase } from './index';
import type { LocalDocument } from './types';

function generateId(): string {
  return `local_${Date.now()}_${Math.random().toString(36).slice(2, 9)}`;
}

/**
 * Upsert a document by server_id.
 */
export async function upsertDocument(doc: Omit<LocalDocument, 'id'>): Promise<string> {
  const db = await getDatabase();

  const existing = await db.getFirstAsync<LocalDocument>(
    'SELECT id FROM documents WHERE server_id = ?',
    [doc.server_id]
  );

  if (existing) {
    await db.runAsync(
      `UPDATE documents SET
        title = ?, icon = ?, locked = ?, lock_type = ?,
        owner_id = ?, owner_name = ?, context_type = ?,
        context_id = ?, updated_at = ?, synced_at = ?
      WHERE server_id = ?`,
      [
        doc.title, doc.icon, doc.locked, doc.lock_type,
        doc.owner_id, doc.owner_name, doc.context_type,
        doc.context_id, doc.updated_at, doc.synced_at,
        doc.server_id,
      ]
    );
    return existing.id;
  }

  const id = generateId();
  await db.runAsync(
    `INSERT INTO documents (id, server_id, title, icon, locked, lock_type,
      owner_id, owner_name, context_type, context_id, updated_at, synced_at)
    VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
    [
      id, doc.server_id, doc.title, doc.icon, doc.locked, doc.lock_type,
      doc.owner_id, doc.owner_name, doc.context_type,
      doc.context_id, doc.updated_at, doc.synced_at,
    ]
  );
  return id;
}

/**
 * Get all documents ordered by update time.
 */
export async function getDocuments(): Promise<LocalDocument[]> {
  const db = await getDatabase();
  return db.getAllAsync<LocalDocument>(
    'SELECT * FROM documents ORDER BY updated_at DESC'
  );
}

/**
 * Get a document by server_id.
 */
export async function getDocumentByServerId(serverId: string): Promise<LocalDocument | null> {
  const db = await getDatabase();
  return db.getFirstAsync<LocalDocument>(
    'SELECT * FROM documents WHERE server_id = ?',
    [serverId]
  );
}

/**
 * Get documents by context (chat or topic).
 */
export async function getDocumentsByContext(
  contextType: string,
  contextId: string
): Promise<LocalDocument[]> {
  const db = await getDatabase();
  return db.getAllAsync<LocalDocument>(
    'SELECT * FROM documents WHERE context_type = ? AND context_id = ? ORDER BY updated_at DESC',
    [contextType, contextId]
  );
}

/**
 * Delete a document.
 */
export async function deleteDocument(serverId: string): Promise<void> {
  const db = await getDatabase();
  await db.runAsync('DELETE FROM documents WHERE server_id = ?', [serverId]);
}

/**
 * Get document count.
 */
export async function getDocumentCount(): Promise<number> {
  const db = await getDatabase();
  const result = await db.getFirstAsync<{ count: number }>(
    'SELECT COUNT(*) as count FROM documents'
  );
  return result?.count ?? 0;
}
