// Local database initialization using expo-sqlite
// Provides a singleton database instance with schema setup

import * as SQLite from 'expo-sqlite';
import { CREATE_TABLES_SQL, DB_VERSION } from './schema';

const DB_NAME = 'chatat.db';

let db: SQLite.SQLiteDatabase | null = null;

/**
 * Get the singleton database instance.
 * Initializes the database and creates tables on first call.
 */
export async function getDatabase(): Promise<SQLite.SQLiteDatabase> {
  if (db) return db;

  db = await SQLite.openDatabaseAsync(DB_NAME);

  // Enable WAL mode for better concurrent read/write performance
  await db.execAsync('PRAGMA journal_mode = WAL;');
  await db.execAsync('PRAGMA foreign_keys = ON;');

  // Check and apply schema
  await applySchema(db);

  return db;
}

/**
 * Apply database schema. Creates tables if they don't exist.
 */
async function applySchema(database: SQLite.SQLiteDatabase): Promise<void> {
  const result = await database.getFirstAsync<{ user_version: number }>(
    'PRAGMA user_version;'
  );
  const currentVersion = result?.user_version ?? 0;

  if (currentVersion < DB_VERSION) {
    await database.execAsync(CREATE_TABLES_SQL);
    await database.execAsync(`PRAGMA user_version = ${DB_VERSION};`);
  }
}

/**
 * Close the database connection.
 * Call this when the app is shutting down.
 */
export async function closeDatabase(): Promise<void> {
  if (db) {
    await db.closeAsync();
    db = null;
  }
}

/**
 * Reset the database — drops all tables and recreates schema.
 * Use only for debugging or user-initiated data clear.
 */
export async function resetDatabase(): Promise<void> {
  const database = await getDatabase();
  await database.execAsync('PRAGMA user_version = 0;');
  db = null;
  // Reopen to recreate
  await getDatabase();
}

export { DB_NAME };
