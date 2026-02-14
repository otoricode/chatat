-- Performance indexes for frequently queried columns

-- Messages: chat_id + created_at (chat message listing, ordered by time)
CREATE INDEX IF NOT EXISTS idx_messages_chat_created
    ON messages (chat_id, created_at DESC);

-- Messages: sender_id (lookup messages by sender)
CREATE INDEX IF NOT EXISTS idx_messages_sender
    ON messages (sender_id);

-- Chat members: user_id (lookup chats a user belongs to)
CREATE INDEX IF NOT EXISTS idx_chat_members_user
    ON chat_members (user_id);

-- Documents: chat_id + created_at (documents tab in chat)
CREATE INDEX IF NOT EXISTS idx_documents_chat_created
    ON documents (chat_id, created_at DESC);

-- Topic messages: topic_id + created_at (topic message listing)
CREATE INDEX IF NOT EXISTS idx_topic_messages_topic_created
    ON topic_messages (topic_id, created_at DESC);

-- Document entities: document_id (list entities linked to a document)
CREATE INDEX IF NOT EXISTS idx_document_entities_doc
    ON document_entities (document_id);

-- Document entities: entity_id (list documents linked to an entity)
CREATE INDEX IF NOT EXISTS idx_document_entities_entity
    ON document_entities (entity_id);
