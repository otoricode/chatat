// SyncEngine — background sync service for offline-first support
// Syncs chats, messages, contacts, and documents between server and local DB.

import { chatsApi } from '@/services/api/chats';
import { contactsApi } from '@/services/api/contacts';
import { documentsApi } from '@/services/api/documents';
import * as chatRepo from '@/database/chatRepo';
import * as messageRepo from '@/database/messageRepo';
import * as contactRepo from '@/database/contactRepo';
import * as documentRepo from '@/database/documentRepo';
import * as syncRepo from '@/database/syncRepo';

type SyncListener = (status: SyncStatus) => void;

export type SyncStatus = {
  isSyncing: boolean;
  lastSyncedAt: number;
  error: string | null;
};

class SyncEngineService {
  private listeners: SyncListener[] = [];
  private isSyncing = false;
  private lastSyncedAt = 0;

  /**
   * Subscribe to sync status changes.
   */
  subscribe(listener: SyncListener): () => void {
    this.listeners.push(listener);
    return () => {
      this.listeners = this.listeners.filter((l) => l !== listener);
    };
  }

  private notify(status: SyncStatus): void {
    for (const listener of this.listeners) {
      listener(status);
    }
  }

  /**
   * Full sync — syncs all entity types.
   * Called on app launch or manual refresh.
   */
  async fullSync(): Promise<void> {
    if (this.isSyncing) return;

    this.isSyncing = true;
    this.notify({ isSyncing: true, lastSyncedAt: this.lastSyncedAt, error: null });

    try {
      await Promise.all([
        this.syncChats(),
        this.syncContacts(),
        this.syncDocuments(),
      ]);

      this.lastSyncedAt = Date.now();
      await syncRepo.setLastSyncTime('full_sync', this.lastSyncedAt);
      this.notify({ isSyncing: false, lastSyncedAt: this.lastSyncedAt, error: null });
    } catch (err) {
      const errorMsg = err instanceof Error ? err.message : 'Sync failed';
      this.notify({ isSyncing: false, lastSyncedAt: this.lastSyncedAt, error: errorMsg });
    } finally {
      this.isSyncing = false;
    }
  }

  /**
   * Sync chats from server to local DB.
   * chatsApi.list() returns ApiResponse<ChatListItem[]> → response.data.data
   */
  async syncChats(): Promise<void> {
    try {
      const response = await chatsApi.list();
      const serverChats = response.data.data;

      for (const item of serverChats) {
        await chatRepo.upsertChat({
          server_id: item.chat.id,
          type: item.chat.type,
          name: item.chat.name ?? '',
          icon: item.chat.icon ?? null,
          last_message: item.lastMessage?.content ?? null,
          last_message_at: item.lastMessage
            ? new Date(item.lastMessage.createdAt).getTime()
            : null,
          unread_count: item.unreadCount,
          is_muted: 0,
          is_archived: 0,
          pinned_at: item.chat.pinnedAt
            ? new Date(item.chat.pinnedAt).getTime()
            : null,
          synced_at: Date.now(),
          created_at: new Date(item.chat.createdAt).getTime(),
        });
      }

      await syncRepo.setLastSyncTime('chats', Date.now());
    } catch {
      // Silently fail — will retry on next sync
    }
  }

  /**
   * Sync messages for a specific chat.
   * Uses cursor-based pagination to fetch only new messages.
   * chatsApi.getMessages() returns PaginatedResponse<Message[]> → response.data.data
   */
  async syncMessages(chatId: string): Promise<void> {
    try {
      const cursor = await messageRepo.getLatestMessageServerId(chatId);
      const response = await chatsApi.getMessages(chatId, cursor ?? undefined, 50);
      const serverMessages = response.data.data;

      for (const msg of serverMessages) {
        await messageRepo.upsertMessage({
          server_id: msg.id,
          chat_id: chatId,
          sender_id: msg.senderId,
          sender_name: '',
          content: msg.content ?? null,
          type: msg.type,
          status: 'delivered',
          reply_to_id: msg.replyToId ?? null,
          metadata: msg.metadata ? JSON.stringify(msg.metadata) : null,
          is_deleted: msg.isDeleted ? 1 : 0,
          is_pending: 0,
          created_at: new Date(msg.createdAt).getTime(),
        });
      }
    } catch {
      // Silently fail
    }
  }

  /**
   * Sync contacts from server to local DB.
   * contactsApi.list() returns Contact[] directly → response.data
   */
  async syncContacts(): Promise<void> {
    try {
      const response = await contactsApi.list();
      const serverContacts = response.data;

      // response.data is Contact[] (no .data wrapper)
      const contacts = Array.isArray(serverContacts) ? serverContacts : [];

      for (const contact of contacts) {
        await contactRepo.upsertContact({
          server_id: contact.id,
          user_id: contact.id,
          name: contact.name,
          phone: contact.phone,
          avatar: contact.avatar ?? null,
          status_text: contact.status ?? null,
          is_registered: 1,
          synced_at: Date.now(),
        });
      }

      await syncRepo.setLastSyncTime('contacts', Date.now());
    } catch {
      // Silently fail
    }
  }

  /**
   * Sync documents from server to local DB.
   * documentsApi.list() returns { data: DocumentListItem[]; meta: ... } → response.data.data
   */
  async syncDocuments(): Promise<void> {
    try {
      const response = await documentsApi.list();
      const serverDocs = response.data.data;

      for (const doc of serverDocs) {
        await documentRepo.upsertDocument({
          server_id: doc.id,
          title: doc.title,
          icon: doc.icon ?? null,
          locked: doc.locked ? 1 : 0,
          lock_type: null,
          owner_id: doc.ownerId ?? '',
          owner_name: '',
          context_type: doc.contextType ?? 'standalone',
          context_id: null,
          updated_at: new Date(doc.updatedAt).getTime(),
          synced_at: Date.now(),
        });
      }

      await syncRepo.setLastSyncTime('documents', Date.now());
    } catch {
      // Silently fail
    }
  }

  /**
   * Get last full sync timestamp.
   */
  async getLastFullSyncTime(): Promise<number> {
    return syncRepo.getLastSyncTime('full_sync');
  }

  /**
   * Check if we need a full sync (e.g., more than 5 minutes since last).
   */
  async needsFullSync(): Promise<boolean> {
    const lastSync = await this.getLastFullSyncTime();
    const fiveMinutes = 5 * 60 * 1000;
    return Date.now() - lastSync > fiveMinutes;
  }
}

// Singleton instance
export const syncEngine = new SyncEngineService();
