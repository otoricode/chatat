import { DB_VERSION, CREATE_TABLES_SQL, DROP_TABLES_SQL } from '../schema';

describe('schema', () => {
  it('has a valid DB_VERSION', () => {
    expect(DB_VERSION).toBe(1);
    expect(typeof DB_VERSION).toBe('number');
  });

  it('CREATE_TABLES_SQL contains all required tables', () => {
    expect(CREATE_TABLES_SQL).toContain('CREATE TABLE IF NOT EXISTS chats');
    expect(CREATE_TABLES_SQL).toContain('CREATE TABLE IF NOT EXISTS messages');
    expect(CREATE_TABLES_SQL).toContain('CREATE TABLE IF NOT EXISTS contacts');
    expect(CREATE_TABLES_SQL).toContain('CREATE TABLE IF NOT EXISTS documents');
    expect(CREATE_TABLES_SQL).toContain('CREATE TABLE IF NOT EXISTS sync_meta');
    expect(CREATE_TABLES_SQL).toContain('CREATE TABLE IF NOT EXISTS pending_operations');
  });

  it('CREATE_TABLES_SQL contains key indexes', () => {
    expect(CREATE_TABLES_SQL).toContain('idx_chats_server_id');
    expect(CREATE_TABLES_SQL).toContain('idx_messages_server_id');
    expect(CREATE_TABLES_SQL).toContain('idx_messages_chat_id');
    expect(CREATE_TABLES_SQL).toContain('idx_contacts_server_id');
    expect(CREATE_TABLES_SQL).toContain('idx_documents_server_id');
    expect(CREATE_TABLES_SQL).toContain('idx_pending_ops_status');
  });

  it('DROP_TABLES_SQL drops all tables', () => {
    expect(DROP_TABLES_SQL).toContain('DROP TABLE IF EXISTS chats');
    expect(DROP_TABLES_SQL).toContain('DROP TABLE IF EXISTS messages');
    expect(DROP_TABLES_SQL).toContain('DROP TABLE IF EXISTS contacts');
    expect(DROP_TABLES_SQL).toContain('DROP TABLE IF EXISTS documents');
    expect(DROP_TABLES_SQL).toContain('DROP TABLE IF EXISTS sync_meta');
    expect(DROP_TABLES_SQL).toContain('DROP TABLE IF EXISTS pending_operations');
  });
});
