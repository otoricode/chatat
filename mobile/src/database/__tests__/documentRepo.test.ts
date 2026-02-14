// @ts-nocheck
const mockDb = {
  runAsync: jest.fn(),
  getAllAsync: jest.fn(),
  getFirstAsync: jest.fn(),
  execAsync: jest.fn(),
};

jest.mock('../index', () => ({
  getDatabase: jest.fn(() => Promise.resolve(mockDb)),
}));

import * as documentRepo from '../documentRepo';

beforeEach(() => {
  jest.clearAllMocks();
});

const baseDoc = {
  server_id: 'doc_srv1',
  title: 'Test Doc',
  icon: null,
  locked: 0,
  lock_type: null,
  owner_id: 'u1',
  owner_name: 'Alice',
  context_type: 'standalone',
  context_id: null,
  updated_at: Date.now(),
  synced_at: Date.now(),
};

describe('documentRepo', () => {
  describe('upsertDocument', () => {
    it('updates when existing document found', async () => {
      mockDb.getFirstAsync.mockResolvedValue({ id: 'existing_id' });
      mockDb.runAsync.mockResolvedValue(undefined);

      const id = await documentRepo.upsertDocument(baseDoc);

      expect(id).toBe('existing_id');
      expect(mockDb.runAsync).toHaveBeenCalledWith(
        expect.stringContaining('UPDATE documents'),
        expect.any(Array),
      );
    });

    it('inserts when no existing document', async () => {
      mockDb.getFirstAsync.mockResolvedValue(null);
      mockDb.runAsync.mockResolvedValue(undefined);

      const id = await documentRepo.upsertDocument(baseDoc);

      expect(id).toMatch(/^local_/);
      expect(mockDb.runAsync).toHaveBeenCalledWith(
        expect.stringContaining('INSERT INTO documents'),
        expect.any(Array),
      );
    });
  });

  describe('getDocuments', () => {
    it('returns all documents', async () => {
      mockDb.getAllAsync.mockResolvedValue([{ id: 'd1' }]);
      const result = await documentRepo.getDocuments();
      expect(result).toHaveLength(1);
      expect(mockDb.getAllAsync).toHaveBeenCalledWith(
        expect.stringContaining('ORDER BY updated_at'),
      );
    });
  });

  describe('getDocumentByServerId', () => {
    it('returns document when found', async () => {
      mockDb.getFirstAsync.mockResolvedValue({ id: 'd1', server_id: 'srv1' });
      const result = await documentRepo.getDocumentByServerId('srv1');
      expect(result).toBeTruthy();
    });

    it('returns null when not found', async () => {
      mockDb.getFirstAsync.mockResolvedValue(null);
      const result = await documentRepo.getDocumentByServerId('nonexistent');
      expect(result).toBeNull();
    });
  });

  describe('getDocumentsByContext', () => {
    it('returns documents by context type and id', async () => {
      mockDb.getAllAsync.mockResolvedValue([{ id: 'd1', context_type: 'chat', context_id: 'c1' }]);
      const result = await documentRepo.getDocumentsByContext('chat', 'c1');
      expect(result).toHaveLength(1);
      expect(mockDb.getAllAsync).toHaveBeenCalledWith(
        expect.stringContaining('context_type = ? AND context_id = ?'),
        ['chat', 'c1'],
      );
    });
  });

  describe('deleteDocument', () => {
    it('deletes by server_id', async () => {
      mockDb.runAsync.mockResolvedValue(undefined);
      await documentRepo.deleteDocument('srv1');
      expect(mockDb.runAsync).toHaveBeenCalledWith(
        expect.stringContaining('DELETE FROM documents'),
        ['srv1'],
      );
    });
  });

  describe('getDocumentCount', () => {
    it('returns count', async () => {
      mockDb.getFirstAsync.mockResolvedValue({ count: 3 });
      const result = await documentRepo.getDocumentCount();
      expect(result).toBe(3);
    });

    it('returns 0 when null', async () => {
      mockDb.getFirstAsync.mockResolvedValue(null);
      const result = await documentRepo.getDocumentCount();
      expect(result).toBe(0);
    });
  });
});
