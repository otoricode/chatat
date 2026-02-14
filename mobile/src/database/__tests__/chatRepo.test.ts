const mockDb = {
  runAsync: jest.fn(),
  getAllAsync: jest.fn(),
  getFirstAsync: jest.fn(),
  execAsync: jest.fn(),
};

jest.mock('../index', () => ({
  getDatabase: jest.fn(() => Promise.resolve(mockDb)),
}));

import * as chatRepo from '../chatRepo';

beforeEach(() => {
  jest.clearAllMocks();
});

const baseChatData = {
  server_id: 'srv1',
  type: 'personal' as const,
  name: 'Alice',
  icon: null,
  last_message: 'Hi',
  last_message_at: 12345,
  unread_count: 0,
  is_muted: 0,
  is_archived: 0,
  pinned_at: null,
  synced_at: Date.now(),
  created_at: Date.now(),
};

describe('chatRepo', () => {
  describe('upsertChat', () => {
    it('updates when existing chat found', async () => {
      mockDb.getFirstAsync.mockResolvedValue({ id: 'existing_id' });
      mockDb.runAsync.mockResolvedValue(undefined);

      const id = await chatRepo.upsertChat(baseChatData);

      expect(id).toBe('existing_id');
      expect(mockDb.runAsync).toHaveBeenCalledWith(
        expect.stringContaining('UPDATE chats'),
        expect.any(Array),
      );
    });

    it('inserts when no existing chat', async () => {
      mockDb.getFirstAsync.mockResolvedValue(null);
      mockDb.runAsync.mockResolvedValue(undefined);

      const id = await chatRepo.upsertChat(baseChatData);

      expect(id).toMatch(/^local_/);
      expect(mockDb.runAsync).toHaveBeenCalledWith(
        expect.stringContaining('INSERT INTO chats'),
        expect.any(Array),
      );
    });
  });

  describe('getChats', () => {
    it('returns all chats ordered', async () => {
      mockDb.getAllAsync.mockResolvedValue([{ id: 'c1', name: 'Alice' }]);

      const result = await chatRepo.getChats();

      expect(result).toHaveLength(1);
      expect(mockDb.getAllAsync).toHaveBeenCalledWith(
        expect.stringContaining('ORDER BY'),
      );
    });
  });

  describe('getChatByServerId', () => {
    it('returns chat when found', async () => {
      mockDb.getFirstAsync.mockResolvedValue({ id: 'c1', server_id: 'srv1' });

      const result = await chatRepo.getChatByServerId('srv1');

      expect(result).toBeTruthy();
      expect(result?.server_id).toBe('srv1');
    });

    it('returns null when not found', async () => {
      mockDb.getFirstAsync.mockResolvedValue(null);

      const result = await chatRepo.getChatByServerId('nonexistent');

      expect(result).toBeNull();
    });
  });

  describe('updateUnreadCount', () => {
    it('updates unread count', async () => {
      mockDb.runAsync.mockResolvedValue(undefined);

      await chatRepo.updateUnreadCount('srv1', 5);

      expect(mockDb.runAsync).toHaveBeenCalledWith(
        expect.stringContaining('UPDATE chats SET unread_count'),
        [5, 'srv1'],
      );
    });
  });

  describe('deleteChat', () => {
    it('deletes messages and chat', async () => {
      mockDb.runAsync.mockResolvedValue(undefined);

      await chatRepo.deleteChat('srv1');

      expect(mockDb.runAsync).toHaveBeenCalledTimes(2);
      expect(mockDb.runAsync).toHaveBeenCalledWith(
        expect.stringContaining('DELETE FROM messages'),
        ['srv1'],
      );
      expect(mockDb.runAsync).toHaveBeenCalledWith(
        expect.stringContaining('DELETE FROM chats'),
        ['srv1'],
      );
    });
  });

  describe('getChatCount', () => {
    it('returns count', async () => {
      mockDb.getFirstAsync.mockResolvedValue({ count: 10 });

      const result = await chatRepo.getChatCount();

      expect(result).toBe(10);
    });

    it('returns 0 when null', async () => {
      mockDb.getFirstAsync.mockResolvedValue(null);

      const result = await chatRepo.getChatCount();

      expect(result).toBe(0);
    });
  });
});
