-- Backup records: tracks user backup history
CREATE TABLE IF NOT EXISTS backup_records (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    size_bytes BIGINT NOT NULL DEFAULT 0,
    platform VARCHAR(20) NOT NULL CHECK (platform IN ('google_drive', 'icloud')),
    status VARCHAR(20) NOT NULL DEFAULT 'completed' CHECK (status IN ('in_progress', 'completed', 'failed')),
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_backup_records_user_id ON backup_records(user_id);
CREATE INDEX idx_backup_records_created_at ON backup_records(created_at DESC);
