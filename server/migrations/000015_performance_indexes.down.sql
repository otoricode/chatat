-- Drop performance indexes

DROP INDEX IF EXISTS idx_document_entities_entity;
DROP INDEX IF EXISTS idx_document_entities_doc;
DROP INDEX IF EXISTS idx_topic_messages_topic_created;
DROP INDEX IF EXISTS idx_documents_chat_created;
DROP INDEX IF EXISTS idx_chat_members_user;
DROP INDEX IF EXISTS idx_messages_sender;
DROP INDEX IF EXISTS idx_messages_chat_created;
