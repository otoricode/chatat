CREATE TABLE topics (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  name VARCHAR(100) NOT NULL,
  icon VARCHAR(10) NOT NULL DEFAULT '💬',
  description TEXT,
  parent_type VARCHAR(10) NOT NULL CHECK(parent_type IN ('personal', 'group')),
  parent_id UUID NOT NULL REFERENCES chats(id) ON DELETE CASCADE,
  created_by UUID NOT NULL REFERENCES users(id),
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE topic_members (
  topic_id UUID NOT NULL REFERENCES topics(id) ON DELETE CASCADE,
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  role VARCHAR(10) NOT NULL DEFAULT 'member'
    CHECK(role IN ('admin', 'member')),
  joined_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  PRIMARY KEY (topic_id, user_id)
);

CREATE INDEX idx_topics_parent_id ON topics(parent_id);
CREATE INDEX idx_topic_members_user_id ON topic_members(user_id);

CREATE TABLE topic_messages (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  topic_id UUID NOT NULL REFERENCES topics(id) ON DELETE CASCADE,
  sender_id UUID NOT NULL REFERENCES users(id),
  content TEXT NOT NULL,
  reply_to_id UUID REFERENCES topic_messages(id) ON DELETE SET NULL,
  type VARCHAR(20) NOT NULL DEFAULT 'text'
    CHECK(type IN ('text', 'image', 'file', 'document_card', 'system')),
  metadata JSONB,
  is_deleted BOOLEAN NOT NULL DEFAULT false,
  deleted_for_all BOOLEAN NOT NULL DEFAULT false,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_topic_messages_topic_id ON topic_messages(topic_id);
CREATE INDEX idx_topic_messages_topic_id_created_at ON topic_messages(topic_id, created_at DESC);
