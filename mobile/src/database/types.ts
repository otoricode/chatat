// Local database row types
// These mirror the SQLite table structures

export type LocalChat = {
  id: string;
  server_id: string;
  type: 'personal' | 'group';
  name: string | null;
  icon: string | null;
  last_message: string | null;
  last_message_at: number | null;
  unread_count: number;
  is_muted: number;
  is_archived: number;
  pinned_at: number | null;
  synced_at: number;
  created_at: number;
};

export type LocalMessage = {
  id: string;
  server_id: string | null;
  chat_id: string;
  sender_id: string;
  sender_name: string;
  content: string | null;
  type: string;
  status: string;
  reply_to_id: string | null;
  metadata: string | null;
  is_deleted: number;
  is_pending: number;
  created_at: number;
};

export type LocalContact = {
  id: string;
  server_id: string;
  user_id: string;
  name: string;
  phone: string;
  avatar: string | null;
  status_text: string | null;
  is_registered: number;
  synced_at: number;
};

export type LocalDocument = {
  id: string;
  server_id: string;
  title: string;
  icon: string | null;
  locked: number;
  lock_type: string | null;
  owner_id: string;
  owner_name: string;
  context_type: string;
  context_id: string | null;
  updated_at: number;
  synced_at: number;
};

export type SyncMeta = {
  key: string;
  value: string;
  updated_at: number;
};

export type PendingOperation = {
  id: string;
  type: string; // 'create' | 'update' | 'delete'
  entity_type: string; // 'message' | 'chat' | 'document'
  entity_id: string;
  payload: string; // JSON
  retry_count: number;
  status: string; // 'pending' | 'processing' | 'failed'
  created_at: number;
  last_attempt_at: number | null;
};
