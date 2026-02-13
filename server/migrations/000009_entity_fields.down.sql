DROP INDEX IF EXISTS idx_entities_owner_type;
DROP INDEX IF EXISTS idx_entities_type;

ALTER TABLE entities DROP COLUMN IF EXISTS updated_at;
ALTER TABLE entities DROP COLUMN IF EXISTS fields;
