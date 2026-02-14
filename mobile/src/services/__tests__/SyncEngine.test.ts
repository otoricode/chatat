// @ts-nocheck
jest.mock('@/services/api/chats', () => ({
  chatsApi: {
    list: jest.fn(),
    getMessages: jest.fn(),
  },
}));

jest.mock('@/services/api/contacts', () => ({
  contactsApi: {
    list: jest.fn(),
  },
}));

jest.mock('@/services/api/documents', () => ({
  documentsApi: {
    list: jest.fn(),
  },
}));

jest.mock('@/database/chatRepo', () => ({
  upsertChat: jest.fn(),
}));

jest.mock('@/database/messageRepo', () => ({
  getLatestMessageServerId: jest.fn(),
  upsertMessage: jest.fn(),
}));

jest.mock('@/database/contactRepo', () => ({
  upsertContact: jest.fn(),
}));

jest.mock('@/database/documentRepo', () => ({
  upsertDocument: jest.fn(),
}));

jest.mock('@/database/syncRepo', () => ({
  setLastSyncTime: jest.fn(),
  getLastSyncTime: jest.fn(),
}));

import { syncEngine } from '../SyncEngine';
import { chatsApi } from '@/services/api/chats';
import { contactsApi } from '@/services/api/contacts';
import { documentsApi } from '@/services/api/documents';
import * as chatRepo from '@/database/chatRepo';
import * as messageRepo from '@/database/messageRepo';
import * as contactRepo from '@/database/contactRepo';
import * as documentRepo from '@/database/documentRepo';
import * as syncRepo from '@/database/syncRepo';

const mockChatsApi = chatsApi as jest.Mocked<typeof chatsApi>;
const mockContactsApi = contactsApi as jest.Mocked<typeof contactsApi>;
const mockDocsApi = documentsApi as jest.Mocked<typeof documentsApi>;
const mockChatRepo = chatRepo as jest.Mocked<typeof chatRepo>;
const mockMessageRepo = messageRepo as jest.Mocked<typeof messageRepo>;
const mockContactRepo = contactRepo as jest.Mocked<typeof contactRepo>;
const mockDocRepo = documentRepo as jest.Mocked<typeof documentRepo>;
const mockSyncRepo = syncRepo as jest.Mocked<typeof syncRepo>;

beforeEach(() => {
  jest.clearAllMocks();
});

describe('SyncEngine', () => {
  describe('subscribe', () => {
    it('returns unsubscribe function', () => {
      const listener = jest.fn();
      const unsub = syncEngine.subscribe(listener);

      expect(typeof unsub).toBe('function');
    });

    it('notifies listener on fullSync', async () => {
      const listener = jest.fn();
      syncEngine.subscribe(listener);

      // Mock all APIs to succeed
      mockChatsApi.list.mockResolvedValue({ data: { data: [] } } as any);
      mockContactsApi.list.mockResolvedValue({ data: [] } as any);
      mockDocsApi.list.mockResolvedValue({ data: { data: [] } } as any);
      mockSyncRepo.setLastSyncTime.mockResolvedValue(undefined);

      await syncEngine.fullSync();

      // Should be called at least twice: starting + completed
      expect(listener).toHaveBeenCalledWith(
        expect.objectContaining({ isSyncing: true }),
      );
      expect(listener).toHaveBeenCalledWith(
        expect.objectContaining({ isSyncing: false }),
      );
    });
  });

  describe('fullSync', () => {
    it('syncs chats, contacts, and documents', async () => {
      mockChatsApi.list.mockResolvedValue({ data: { data: [] } } as any);
      mockContactsApi.list.mockResolvedValue({ data: [] } as any);
      mockDocsApi.list.mockResolvedValue({ data: { data: [] } } as any);
      mockSyncRepo.setLastSyncTime.mockResolvedValue(undefined);

      await syncEngine.fullSync();

      expect(mockChatsApi.list).toHaveBeenCalled();
      expect(mockContactsApi.list).toHaveBeenCalled();
      expect(mockDocsApi.list).toHaveBeenCalled();
    });

    it('handles error during sync silently in individual methods', async () => {
      const listener = jest.fn();
      syncEngine.subscribe(listener);

      // Individual sync methods catch errors silently,
      // so fullSync completes without error
      mockChatsApi.list.mockRejectedValue(new Error('Network error'));
      mockContactsApi.list.mockResolvedValue({ data: [] } as any);
      mockDocsApi.list.mockResolvedValue({ data: { data: [] } } as any);
      mockSyncRepo.setLastSyncTime.mockResolvedValue(undefined);

      await syncEngine.fullSync();

      // Should complete successfully since individual methods catch errors
      expect(listener).toHaveBeenCalledWith(
        expect.objectContaining({ isSyncing: false }),
      );
    });

    it('completes sync when all individual syncs fail silently', async () => {
      const listener = jest.fn();
      syncEngine.subscribe(listener);

      mockChatsApi.list.mockRejectedValue('fail');
      mockContactsApi.list.mockRejectedValue('fail');
      mockDocsApi.list.mockRejectedValue('fail');
      mockSyncRepo.setLastSyncTime.mockResolvedValue(undefined);

      await syncEngine.fullSync();

      // Individual sync methods catch their own errors, so fullSync still completes
      expect(listener).toHaveBeenCalledWith(
        expect.objectContaining({ isSyncing: false }),
      );
    });
  });

  describe('syncChats', () => {
    it('upserts chats from server', async () => {
      const chatItems = [
        {
          chat: { id: 'c1', type: 'personal', name: 'Alice', icon: null, pinnedAt: null, createdAt: '2024-01-01T00:00:00Z' },
          lastMessage: { content: 'Hi', createdAt: '2024-01-01T00:00:00Z' },
          unreadCount: 2,
        },
      ];
      mockChatsApi.list.mockResolvedValue({ data: { data: chatItems } } as any);
      mockSyncRepo.setLastSyncTime.mockResolvedValue(undefined);

      await syncEngine.syncChats();

      expect(mockChatRepo.upsertChat).toHaveBeenCalledWith(
        expect.objectContaining({ server_id: 'c1', name: 'Alice', unread_count: 2 }),
      );
      expect(mockSyncRepo.setLastSyncTime).toHaveBeenCalledWith('chats', expect.any(Number));
    });

    it('handles chat with no lastMessage', async () => {
      const chatItems = [
        {
          chat: { id: 'c1', type: 'personal', name: 'Alice', icon: null, pinnedAt: null, createdAt: '2024-01-01T00:00:00Z' },
          lastMessage: null,
          unreadCount: 0,
        },
      ];
      mockChatsApi.list.mockResolvedValue({ data: { data: chatItems } } as any);
      mockSyncRepo.setLastSyncTime.mockResolvedValue(undefined);

      await syncEngine.syncChats();

      expect(mockChatRepo.upsertChat).toHaveBeenCalledWith(
        expect.objectContaining({ last_message: null, last_message_at: null }),
      );
    });

    it('handles error silently', async () => {
      mockChatsApi.list.mockRejectedValue(new Error('fail'));

      await syncEngine.syncChats();

      // No throw
      expect(mockChatRepo.upsertChat).not.toHaveBeenCalled();
    });
  });

  describe('syncMessages', () => {
    it('upserts messages from server', async () => {
      mockMessageRepo.getLatestMessageServerId.mockResolvedValue('msg0' as any);
      const msgs = [
        {
          id: 'msg1',
          senderId: 'u1',
          content: 'Hello',
          type: 'text',
          createdAt: '2024-01-01T00:00:00Z',
          isDeleted: false,
          replyToId: null,
          metadata: null,
        },
      ];
      mockChatsApi.getMessages.mockResolvedValue({ data: { data: msgs } } as any);

      await syncEngine.syncMessages('c1');

      expect(mockMessageRepo.upsertMessage).toHaveBeenCalledWith(
        expect.objectContaining({ server_id: 'msg1', chat_id: 'c1' }),
      );
    });

    it('handles no cursor', async () => {
      mockMessageRepo.getLatestMessageServerId.mockResolvedValue(null as any);
      mockChatsApi.getMessages.mockResolvedValue({ data: { data: [] } } as any);

      await syncEngine.syncMessages('c1');

      expect(mockChatsApi.getMessages).toHaveBeenCalled();
    });

    it('handles error silently', async () => {
      mockMessageRepo.getLatestMessageServerId.mockRejectedValue(new Error('DB fail'));

      await syncEngine.syncMessages('c1');

      // No throw
    });
  });

  describe('syncContacts', () => {
    it('upserts contacts from server', async () => {
      const contacts = [
        { id: 'u1', name: 'Alice', phone: '+62111', avatar: null, status: 'Hello' },
      ];
      mockContactsApi.list.mockResolvedValue({ data: contacts } as any);
      mockSyncRepo.setLastSyncTime.mockResolvedValue(undefined);

      await syncEngine.syncContacts();

      expect(mockContactRepo.upsertContact).toHaveBeenCalledWith(
        expect.objectContaining({ server_id: 'u1', name: 'Alice' }),
      );
    });

    it('handles non-array response', async () => {
      mockContactsApi.list.mockResolvedValue({ data: { success: true } } as any);
      mockSyncRepo.setLastSyncTime.mockResolvedValue(undefined);

      await syncEngine.syncContacts();

      // Should not crash
      expect(mockContactRepo.upsertContact).not.toHaveBeenCalled();
    });

    it('handles error silently', async () => {
      mockContactsApi.list.mockRejectedValue(new Error('fail'));

      await syncEngine.syncContacts();

      // No throw
    });
  });

  describe('syncDocuments', () => {
    it('upserts documents from server', async () => {
      const docs = [
        {
          id: 'd1',
          title: 'Doc 1',
          icon: '📄',
          locked: false,
          ownerId: 'u1',
          contextType: 'standalone',
          updatedAt: '2024-01-01T00:00:00Z',
        },
      ];
      mockDocsApi.list.mockResolvedValue({ data: { data: docs } } as any);
      mockSyncRepo.setLastSyncTime.mockResolvedValue(undefined);

      await syncEngine.syncDocuments();

      expect(mockDocRepo.upsertDocument).toHaveBeenCalledWith(
        expect.objectContaining({ server_id: 'd1', title: 'Doc 1' }),
      );
    });

    it('handles error silently', async () => {
      mockDocsApi.list.mockRejectedValue(new Error('fail'));

      await syncEngine.syncDocuments();

      // No throw
    });
  });

  describe('getLastFullSyncTime', () => {
    it('returns last sync time', async () => {
      mockSyncRepo.getLastSyncTime.mockResolvedValue(12345 as any);

      const result = await syncEngine.getLastFullSyncTime();

      expect(result).toBe(12345);
      expect(mockSyncRepo.getLastSyncTime).toHaveBeenCalledWith('full_sync');
    });
  });

  describe('needsFullSync', () => {
    it('returns true when last sync was long ago', async () => {
      mockSyncRepo.getLastSyncTime.mockResolvedValue(0 as any);

      const result = await syncEngine.needsFullSync();

      expect(result).toBe(true);
    });

    it('returns false when synced recently', async () => {
      mockSyncRepo.getLastSyncTime.mockResolvedValue(Date.now() as any);

      const result = await syncEngine.needsFullSync();

      expect(result).toBe(false);
    });
  });
});
