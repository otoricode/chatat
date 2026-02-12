DROP INDEX IF EXISTS idx_users_phone_hash;
ALTER TABLE users DROP COLUMN IF EXISTS phone_hash;
DROP INDEX IF EXISTS idx_user_contacts_contact_user_id;
DROP INDEX IF EXISTS idx_user_contacts_user_id;
DROP TABLE IF EXISTS user_contacts;
