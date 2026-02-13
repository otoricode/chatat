// Sync metadata repository — tracks sync timestamps and pending operations
import { getDatabase } from './index';
import type { PendingOperation } from './types';

function generateId(): string {
  return `op_${Date.now()}_${Math.random().toString(36).slice(2, 9)}`;
}

// --- Sync Timestamps ---

/**
 * Get the last sync timestamp for a given entity type.
 */
export async function getLastSyncTime(key: string): Promise<number> {
  const db = await getDatabase();
  const result = await db.getFirstAsync<{ value: string }>(
    'SELECT value FROM sync_meta WHERE key = ?',
    [key]
  );
  return result ? parseInt(result.value, 10) : 0;
}

/**
 * Set the last sync timestamp for a given entity type.
 */
export async function setLastSyncTime(key: string, timestamp: number): Promise<void> {
  const db = await getDatabase();
  await db.runAsync(
    `INSERT INTO sync_meta (key, value, updated_at)
     VALUES (?, ?, ?)
     ON CONFLICT(key) DO UPDATE SET value = ?, updated_at = ?`,
    [key, String(timestamp), Date.now(), String(timestamp), Date.now()]
  );
}

// --- Pending Operations ---

/**
 * Add a pending operation to the queue.
 */
export async function addPendingOperation(
  type: string,
  entityType: string,
  entityId: string,
  payload: Record<string, unknown>
): Promise<string> {
  const db = await getDatabase();
  const id = generateId();

  await db.runAsync(
    `INSERT INTO pending_operations (id, type, entity_type, entity_id, payload, retry_count, status, created_at)
     VALUES (?, ?, ?, ?, ?, 0, 'pending', ?)`,
    [id, type, entityType, entityId, JSON.stringify(payload), Date.now()]
  );
  return id;
}

/**
 * Get all pending operations, ordered by creation time.
 */
export async function getPendingOperations(): Promise<PendingOperation[]> {
  const db = await getDatabase();
  return db.getAllAsync<PendingOperation>(
    "SELECT * FROM pending_operations WHERE status = 'pending' ORDER BY created_at ASC"
  );
}

/**
 * Mark a pending operation as processing.
 */
export async function markOperationProcessing(id: string): Promise<void> {
  const db = await getDatabase();
  await db.runAsync(
    "UPDATE pending_operations SET status = 'processing', last_attempt_at = ? WHERE id = ?",
    [Date.now(), id]
  );
}

/**
 * Mark a pending operation as completed (remove it).
 */
export async function removeOperation(id: string): Promise<void> {
  const db = await getDatabase();
  await db.runAsync('DELETE FROM pending_operations WHERE id = ?', [id]);
}

/**
 * Mark a pending operation as failed and increment retry count.
 */
export async function markOperationFailed(id: string): Promise<void> {
  const db = await getDatabase();
  await db.runAsync(
    "UPDATE pending_operations SET status = 'pending', retry_count = retry_count + 1, last_attempt_at = ? WHERE id = ?",
    [Date.now(), id]
  );
}

/**
 * Permanently fail an operation (max retries exceeded).
 */
export async function markOperationPermanentlyFailed(id: string): Promise<void> {
  const db = await getDatabase();
  await db.runAsync(
    "UPDATE pending_operations SET status = 'failed', last_attempt_at = ? WHERE id = ?",
    [Date.now(), id]
  );
}

/**
 * Get count of pending operations.
 */
export async function getPendingOperationCount(): Promise<number> {
  const db = await getDatabase();
  const result = await db.getFirstAsync<{ count: number }>(
    "SELECT COUNT(*) as count FROM pending_operations WHERE status = 'pending'"
  );
  return result?.count ?? 0;
}

/**
 * Reset all processing operations back to pending.
 * Call this on app restart to recover from crashes.
 */
export async function resetProcessingOperations(): Promise<void> {
  const db = await getDatabase();
  await db.runAsync(
    "UPDATE pending_operations SET status = 'pending' WHERE status = 'processing'"
  );
}
