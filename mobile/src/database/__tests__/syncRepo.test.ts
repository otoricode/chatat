const mockDb = {
  runAsync: jest.fn(),
  getAllAsync: jest.fn(),
  getFirstAsync: jest.fn(),
  execAsync: jest.fn(),
};

jest.mock('../index', () => ({
  getDatabase: jest.fn(() => Promise.resolve(mockDb)),
}));

import * as syncRepo from '../syncRepo';

beforeEach(() => {
  jest.clearAllMocks();
});

describe('syncRepo', () => {
  describe('getLastSyncTime', () => {
    it('returns parsed timestamp', async () => {
      mockDb.getFirstAsync.mockResolvedValue({ value: '1234567890' });
      const result = await syncRepo.getLastSyncTime('messages');
      expect(result).toBe(1234567890);
    });

    it('returns 0 when no result', async () => {
      mockDb.getFirstAsync.mockResolvedValue(null);
      const result = await syncRepo.getLastSyncTime('messages');
      expect(result).toBe(0);
    });
  });

  describe('setLastSyncTime', () => {
    it('upserts sync time', async () => {
      mockDb.runAsync.mockResolvedValue(undefined);
      await syncRepo.setLastSyncTime('messages', 1234567890);
      expect(mockDb.runAsync).toHaveBeenCalledWith(
        expect.stringContaining('INSERT INTO sync_meta'),
        expect.arrayContaining(['messages', '1234567890']),
      );
    });
  });

  describe('addPendingOperation', () => {
    it('inserts and returns id', async () => {
      mockDb.runAsync.mockResolvedValue(undefined);
      const id = await syncRepo.addPendingOperation('create', 'message', 'e1', { text: 'hi' });
      expect(id).toMatch(/^op_/);
      expect(mockDb.runAsync).toHaveBeenCalledWith(
        expect.stringContaining('INSERT INTO pending_operations'),
        expect.arrayContaining(['create', 'message', 'e1']),
      );
    });
  });

  describe('getPendingOperations', () => {
    it('returns pending operations', async () => {
      mockDb.getAllAsync.mockResolvedValue([{ id: 'op1', status: 'pending' }]);
      const result = await syncRepo.getPendingOperations();
      expect(result).toHaveLength(1);
      expect(mockDb.getAllAsync).toHaveBeenCalledWith(
        expect.stringContaining("status = 'pending'"),
      );
    });
  });

  describe('markOperationProcessing', () => {
    it('updates status to processing', async () => {
      mockDb.runAsync.mockResolvedValue(undefined);
      await syncRepo.markOperationProcessing('op1');
      expect(mockDb.runAsync).toHaveBeenCalledWith(
        expect.stringContaining("status = 'processing'"),
        expect.arrayContaining(['op1']),
      );
    });
  });

  describe('removeOperation', () => {
    it('deletes operation', async () => {
      mockDb.runAsync.mockResolvedValue(undefined);
      await syncRepo.removeOperation('op1');
      expect(mockDb.runAsync).toHaveBeenCalledWith(
        expect.stringContaining('DELETE FROM pending_operations'),
        ['op1'],
      );
    });
  });

  describe('markOperationFailed', () => {
    it('resets to pending with retry increment', async () => {
      mockDb.runAsync.mockResolvedValue(undefined);
      await syncRepo.markOperationFailed('op1');
      expect(mockDb.runAsync).toHaveBeenCalledWith(
        expect.stringContaining("status = 'pending'"),
        expect.arrayContaining(['op1']),
      );
    });
  });

  describe('markOperationPermanentlyFailed', () => {
    it('sets status to failed', async () => {
      mockDb.runAsync.mockResolvedValue(undefined);
      await syncRepo.markOperationPermanentlyFailed('op1');
      expect(mockDb.runAsync).toHaveBeenCalledWith(
        expect.stringContaining("status = 'failed'"),
        expect.arrayContaining(['op1']),
      );
    });
  });

  describe('getPendingOperationCount', () => {
    it('returns count', async () => {
      mockDb.getFirstAsync.mockResolvedValue({ count: 3 });
      const result = await syncRepo.getPendingOperationCount();
      expect(result).toBe(3);
    });

    it('returns 0 when null', async () => {
      mockDb.getFirstAsync.mockResolvedValue(null);
      const result = await syncRepo.getPendingOperationCount();
      expect(result).toBe(0);
    });
  });

  describe('resetProcessingOperations', () => {
    it('resets processing to pending', async () => {
      mockDb.runAsync.mockResolvedValue(undefined);
      await syncRepo.resetProcessingOperations();
      expect(mockDb.runAsync).toHaveBeenCalledWith(
        expect.stringContaining("status = 'pending'"),
      );
    });
  });
});
