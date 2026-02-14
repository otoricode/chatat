const mockDb = {
  runAsync: jest.fn(),
  getAllAsync: jest.fn(),
  getFirstAsync: jest.fn(),
  execAsync: jest.fn(),
};

jest.mock('../index', () => ({
  getDatabase: jest.fn(() => Promise.resolve(mockDb)),
}));

import * as messageRepo from '../messageRepo';

beforeEach(() => {
  jest.clearAllMocks();
});

describe('messageRepo', () => {
  describe('insertMessage', () => {
    it('inserts and returns local id', async () => {
      mockDb.runAsync.mockResolvedValue(undefined);

      const id = await messageRepo.insertMessage({
        server_id: null,
        chat_id: 'c1',
        sender_id: 'u1',
        sender_name: 'Alice',
        content: 'Hello',
        type: 'text',
        status: 'sending',
        reply_to_id: null,
        metadata: null,
        is_deleted: 0,
        is_pending: 1,
        created_at: Date.now(),
      });

      expect(id).toMatch(/^local_/);
      expect(mockDb.runAsync).toHaveBeenCalled();
    });
  });

  describe('upsertMessage', () => {
    it('inserts when no server_id', async () => {
      mockDb.runAsync.mockResolvedValue(undefined);

      const id = await messageRepo.upsertMessage({
        server_id: null,
        chat_id: 'c1',
        sender_id: 'u1',
        sender_name: 'Alice',
        content: 'Hi',
        type: 'text',
        status: 'sent',
        reply_to_id: null,
        metadata: null,
        is_deleted: 0,
        is_pending: 0,
        created_at: Date.now(),
      });

      expect(id).toMatch(/^local_/);
    });

    it('updates when server_id exists in DB', async () => {
      mockDb.getFirstAsync.mockResolvedValue({ id: 'existing_local' });
      mockDb.runAsync.mockResolvedValue(undefined);

      const id = await messageRepo.upsertMessage({
        server_id: 'srv1',
        chat_id: 'c1',
        sender_id: 'u1',
        sender_name: 'Alice',
        content: 'Updated',
        type: 'text',
        status: 'delivered',
        reply_to_id: null,
        metadata: null,
        is_deleted: 0,
        is_pending: 0,
        created_at: Date.now(),
      });

      expect(id).toBe('existing_local');
      expect(mockDb.runAsync).toHaveBeenCalledWith(
        expect.stringContaining('UPDATE messages'),
        expect.any(Array),
      );
    });

    it('inserts when server_id not found in DB', async () => {
      mockDb.getFirstAsync.mockResolvedValue(null);
      mockDb.runAsync.mockResolvedValue(undefined);

      const id = await messageRepo.upsertMessage({
        server_id: 'srv_new',
        chat_id: 'c1',
        sender_id: 'u1',
        sender_name: 'Alice',
        content: 'New',
        type: 'text',
        status: 'sent',
        reply_to_id: null,
        metadata: null,
        is_deleted: 0,
        is_pending: 0,
        created_at: Date.now(),
      });

      expect(id).toMatch(/^local_/);
    });
  });

  describe('getMessages', () => {
    it('gets messages without timestamp', async () => {
      mockDb.getAllAsync.mockResolvedValue([{ id: 'msg1' }]);

      const result = await messageRepo.getMessages('c1');

      expect(result).toHaveLength(1);
      expect(mockDb.getAllAsync).toHaveBeenCalledWith(
        expect.stringContaining('ORDER BY created_at DESC'),
        expect.arrayContaining(['c1', 50]),
      );
    });

    it('gets messages with beforeTimestamp', async () => {
      mockDb.getAllAsync.mockResolvedValue([]);

      await messageRepo.getMessages('c1', 20, 12345);

      expect(mockDb.getAllAsync).toHaveBeenCalledWith(
        expect.stringContaining('created_at <'),
        expect.arrayContaining(['c1', 12345, 20]),
      );
    });
  });

  describe('getPendingMessages', () => {
    it('returns pending messages', async () => {
      mockDb.getAllAsync.mockResolvedValue([{ id: 'p1', is_pending: 1 }]);

      const result = await messageRepo.getPendingMessages();

      expect(result).toHaveLength(1);
      expect(mockDb.getAllAsync).toHaveBeenCalledWith(
        expect.stringContaining('is_pending = 1'),
      );
    });
  });

  describe('getPendingMessagesForChat', () => {
    it('returns pending messages for chat', async () => {
      mockDb.getAllAsync.mockResolvedValue([]);

      await messageRepo.getPendingMessagesForChat('c1');

      expect(mockDb.getAllAsync).toHaveBeenCalledWith(
        expect.stringContaining('is_pending = 1'),
        ['c1'],
      );
    });
  });

  describe('markMessageSent', () => {
    it('updates status and server_id', async () => {
      mockDb.runAsync.mockResolvedValue(undefined);

      await messageRepo.markMessageSent('local_1', 'server_1');

      expect(mockDb.runAsync).toHaveBeenCalledWith(
        expect.stringContaining('UPDATE messages'),
        ['server_1', 'sent', 'local_1'],
      );
    });
  });

  describe('markMessageFailed', () => {
    it('updates status to failed', async () => {
      mockDb.runAsync.mockResolvedValue(undefined);

      await messageRepo.markMessageFailed('local_1');

      expect(mockDb.runAsync).toHaveBeenCalledWith(
        expect.stringContaining('failed'),
        ['local_1'],
      );
    });
  });

  describe('updateMessageStatus', () => {
    it('updates status by server_id', async () => {
      mockDb.runAsync.mockResolvedValue(undefined);

      await messageRepo.updateMessageStatus('srv1', 'read');

      expect(mockDb.runAsync).toHaveBeenCalledWith(
        expect.stringContaining('UPDATE messages SET status'),
        ['read', 'srv1'],
      );
    });
  });

  describe('softDeleteMessage', () => {
    it('soft deletes by server_id', async () => {
      mockDb.runAsync.mockResolvedValue(undefined);

      await messageRepo.softDeleteMessage('srv1');

      expect(mockDb.runAsync).toHaveBeenCalledWith(
        expect.stringContaining('is_deleted = 1'),
        ['srv1'],
      );
    });
  });

  describe('getLatestMessageServerId', () => {
    it('returns server_id of latest message', async () => {
      mockDb.getFirstAsync.mockResolvedValue({ server_id: 'srv_latest' });

      const result = await messageRepo.getLatestMessageServerId('c1');

      expect(result).toBe('srv_latest');
    });

    it('returns null when no messages', async () => {
      mockDb.getFirstAsync.mockResolvedValue(null);

      const result = await messageRepo.getLatestMessageServerId('c1');

      expect(result).toBeNull();
    });
  });

  describe('getMessageCount', () => {
    it('returns count', async () => {
      mockDb.getFirstAsync.mockResolvedValue({ count: 42 });

      const result = await messageRepo.getMessageCount('c1');

      expect(result).toBe(42);
    });

    it('returns 0 when null', async () => {
      mockDb.getFirstAsync.mockResolvedValue(null);

      const result = await messageRepo.getMessageCount('c1');

      expect(result).toBe(0);
    });
  });

  describe('deleteMessagesForChat', () => {
    it('deletes all messages for chat', async () => {
      mockDb.runAsync.mockResolvedValue(undefined);

      await messageRepo.deleteMessagesForChat('c1');

      expect(mockDb.runAsync).toHaveBeenCalledWith(
        expect.stringContaining('DELETE FROM messages'),
        ['c1'],
      );
    });
  });
});
