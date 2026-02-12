CREATE TABLE user_contacts (
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  contact_user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  contact_name VARCHAR(100),
  synced_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  PRIMARY KEY (user_id, contact_user_id)
);

CREATE INDEX idx_user_contacts_user_id ON user_contacts(user_id);
CREATE INDEX idx_user_contacts_contact_user_id ON user_contacts(contact_user_id);

-- Add phone_hash column for fast hash-based lookups during contact sync
ALTER TABLE users ADD COLUMN phone_hash VARCHAR(64);
CREATE INDEX idx_users_phone_hash ON users(phone_hash);
