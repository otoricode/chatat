CREATE TABLE media (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    uploader_id UUID NOT NULL REFERENCES users(id),
    type VARCHAR(10) NOT NULL CHECK(type IN ('image', 'file')),
    filename VARCHAR(255) NOT NULL,
    content_type VARCHAR(100) NOT NULL,
    size INTEGER NOT NULL,
    width INTEGER,
    height INTEGER,
    storage_key VARCHAR(500) NOT NULL,
    thumbnail_key VARCHAR(500),
    context_type VARCHAR(20) CHECK(context_type IN ('chat', 'topic', 'document')),
    context_id UUID,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_media_uploader_id ON media(uploader_id);
CREATE INDEX idx_media_context ON media(context_type, context_id);
