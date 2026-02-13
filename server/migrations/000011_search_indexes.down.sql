-- Drop triggers
DROP TRIGGER IF EXISTS messages_search_trigger ON messages;
DROP TRIGGER IF EXISTS documents_search_trigger ON documents;
DROP TRIGGER IF EXISTS blocks_search_trigger ON blocks;
DROP TRIGGER IF EXISTS users_search_trigger ON users;
DROP TRIGGER IF EXISTS entities_search_trigger ON entities;

-- Drop functions
DROP FUNCTION IF EXISTS messages_search_update();
DROP FUNCTION IF EXISTS documents_search_update();
DROP FUNCTION IF EXISTS blocks_search_update();
DROP FUNCTION IF EXISTS users_search_update();
DROP FUNCTION IF EXISTS entities_search_update();

-- Drop indexes
DROP INDEX IF EXISTS idx_messages_search;
DROP INDEX IF EXISTS idx_documents_search;
DROP INDEX IF EXISTS idx_blocks_search;
DROP INDEX IF EXISTS idx_users_search;
DROP INDEX IF EXISTS idx_entities_search;

-- Drop columns
ALTER TABLE messages DROP COLUMN IF EXISTS search_vector;
ALTER TABLE documents DROP COLUMN IF EXISTS search_vector;
ALTER TABLE blocks DROP COLUMN IF EXISTS search_vector;
ALTER TABLE users DROP COLUMN IF EXISTS search_vector;
ALTER TABLE entities DROP COLUMN IF EXISTS search_vector;
