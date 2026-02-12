CREATE TABLE documents (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  title VARCHAR(200) NOT NULL,
  icon VARCHAR(10) NOT NULL DEFAULT '📄',
  cover VARCHAR(50),
  owner_id UUID NOT NULL REFERENCES users(id),
  chat_id UUID REFERENCES chats(id) ON DELETE SET NULL,
  topic_id UUID REFERENCES topics(id) ON DELETE SET NULL,
  is_standalone BOOLEAN NOT NULL DEFAULT false,
  require_sigs BOOLEAN NOT NULL DEFAULT false,
  locked BOOLEAN NOT NULL DEFAULT false,
  locked_at TIMESTAMPTZ,
  locked_by VARCHAR(20) CHECK(locked_by IN ('manual', 'signatures')),
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE document_collaborators (
  document_id UUID NOT NULL REFERENCES documents(id) ON DELETE CASCADE,
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  role VARCHAR(10) NOT NULL DEFAULT 'editor'
    CHECK(role IN ('editor', 'viewer')),
  added_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  PRIMARY KEY (document_id, user_id)
);

CREATE TABLE document_signers (
  document_id UUID NOT NULL REFERENCES documents(id) ON DELETE CASCADE,
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  signed_at TIMESTAMPTZ,
  signer_name VARCHAR(100),
  PRIMARY KEY (document_id, user_id)
);

CREATE TABLE document_tags (
  document_id UUID NOT NULL REFERENCES documents(id) ON DELETE CASCADE,
  tag VARCHAR(100) NOT NULL,
  PRIMARY KEY (document_id, tag)
);

CREATE TABLE blocks (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  document_id UUID NOT NULL REFERENCES documents(id) ON DELETE CASCADE,
  type VARCHAR(30) NOT NULL
    CHECK(type IN (
      'paragraph', 'heading1', 'heading2', 'heading3',
      'bullet-list', 'numbered-list', 'checklist',
      'table', 'callout', 'code', 'toggle', 'divider', 'quote'
    )),
  content TEXT,
  checked BOOLEAN,
  rows JSONB,
  columns JSONB,
  language VARCHAR(30),
  emoji VARCHAR(10),
  color VARCHAR(20),
  sort_order INTEGER NOT NULL DEFAULT 0,
  parent_block_id UUID REFERENCES blocks(id) ON DELETE CASCADE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE document_history (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  document_id UUID NOT NULL REFERENCES documents(id) ON DELETE CASCADE,
  user_id UUID NOT NULL REFERENCES users(id),
  action VARCHAR(50) NOT NULL,
  details JSONB,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_documents_owner_id ON documents(owner_id);
CREATE INDEX idx_documents_chat_id ON documents(chat_id);
CREATE INDEX idx_documents_topic_id ON documents(topic_id);
CREATE INDEX idx_documents_locked ON documents(locked);
CREATE INDEX idx_blocks_document_id ON blocks(document_id);
CREATE INDEX idx_blocks_sort_order ON blocks(document_id, sort_order);
CREATE INDEX idx_document_history_document_id ON document_history(document_id);
CREATE INDEX idx_document_tags_tag ON document_tags(tag);
