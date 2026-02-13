-- Full-text search indexes using tsvector/GIN

-- Messages FTS
ALTER TABLE messages ADD COLUMN IF NOT EXISTS search_vector tsvector;
CREATE INDEX IF NOT EXISTS idx_messages_search ON messages USING GIN(search_vector);

CREATE OR REPLACE FUNCTION messages_search_update() RETURNS trigger AS $$
BEGIN
  NEW.search_vector := to_tsvector('indonesian', COALESCE(NEW.content, ''));
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER messages_search_trigger
  BEFORE INSERT OR UPDATE OF content ON messages
  FOR EACH ROW EXECUTE FUNCTION messages_search_update();

-- Backfill existing messages
UPDATE messages SET search_vector = to_tsvector('indonesian', COALESCE(content, ''))
WHERE search_vector IS NULL;

-- Documents FTS (title)
ALTER TABLE documents ADD COLUMN IF NOT EXISTS search_vector tsvector;
CREATE INDEX IF NOT EXISTS idx_documents_search ON documents USING GIN(search_vector);

CREATE OR REPLACE FUNCTION documents_search_update() RETURNS trigger AS $$
BEGIN
  NEW.search_vector := to_tsvector('indonesian', COALESCE(NEW.title, ''));
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER documents_search_trigger
  BEFORE INSERT OR UPDATE OF title ON documents
  FOR EACH ROW EXECUTE FUNCTION documents_search_update();

UPDATE documents SET search_vector = to_tsvector('indonesian', COALESCE(title, ''))
WHERE search_vector IS NULL;

-- Blocks FTS (content)
ALTER TABLE blocks ADD COLUMN IF NOT EXISTS search_vector tsvector;
CREATE INDEX IF NOT EXISTS idx_blocks_search ON blocks USING GIN(search_vector);

CREATE OR REPLACE FUNCTION blocks_search_update() RETURNS trigger AS $$
BEGIN
  NEW.search_vector := to_tsvector('indonesian', COALESCE(NEW.content, ''));
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER blocks_search_trigger
  BEFORE INSERT OR UPDATE OF content ON blocks
  FOR EACH ROW EXECUTE FUNCTION blocks_search_update();

UPDATE blocks SET search_vector = to_tsvector('indonesian', COALESCE(content, ''))
WHERE search_vector IS NULL;

-- Users FTS (name, phone, status)
ALTER TABLE users ADD COLUMN IF NOT EXISTS search_vector tsvector;
CREATE INDEX IF NOT EXISTS idx_users_search ON users USING GIN(search_vector);

CREATE OR REPLACE FUNCTION users_search_update() RETURNS trigger AS $$
BEGIN
  NEW.search_vector := to_tsvector('simple',
    COALESCE(NEW.name, '') || ' ' || COALESCE(NEW.phone, '') || ' ' || COALESCE(NEW.status, '')
  );
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER users_search_trigger
  BEFORE INSERT OR UPDATE OF name, phone, status ON users
  FOR EACH ROW EXECUTE FUNCTION users_search_update();

UPDATE users SET search_vector = to_tsvector('simple',
  COALESCE(name, '') || ' ' || COALESCE(phone, '') || ' ' || COALESCE(status, '')
)
WHERE search_vector IS NULL;

-- Entities FTS (name, type)
ALTER TABLE entities ADD COLUMN IF NOT EXISTS search_vector tsvector;
CREATE INDEX IF NOT EXISTS idx_entities_search ON entities USING GIN(search_vector);

CREATE OR REPLACE FUNCTION entities_search_update() RETURNS trigger AS $$
BEGIN
  NEW.search_vector := to_tsvector('simple',
    COALESCE(NEW.name, '') || ' ' || COALESCE(NEW.type, '')
  );
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER entities_search_trigger
  BEFORE INSERT OR UPDATE OF name, type ON entities
  FOR EACH ROW EXECUTE FUNCTION entities_search_update();

UPDATE entities SET search_vector = to_tsvector('simple',
  COALESCE(name, '') || ' ' || COALESCE(type, '')
)
WHERE search_vector IS NULL;
