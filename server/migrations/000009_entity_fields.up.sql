ALTER TABLE entities ADD COLUMN fields JSONB DEFAULT '{}'::jsonb;
ALTER TABLE entities ADD COLUMN updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW();

CREATE INDEX idx_entities_type ON entities(type);
CREATE INDEX idx_entities_owner_type ON entities(owner_id, type);
