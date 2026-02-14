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

import * as contactRepo from '../contactRepo';

beforeEach(() => {
  jest.clearAllMocks();
});

const baseContact = {
  server_id: 'srv1',
  user_id: 'u1',
  name: 'Alice',
  phone: '+628123',
  avatar: null,
  status_text: 'Hi there',
  is_registered: 1,
  synced_at: Date.now(),
};

describe('contactRepo', () => {
  describe('upsertContact', () => {
    it('updates when existing contact found', async () => {
      mockDb.getFirstAsync.mockResolvedValue({ id: 'existing_id' });
      mockDb.runAsync.mockResolvedValue(undefined);

      const id = await contactRepo.upsertContact(baseContact);

      expect(id).toBe('existing_id');
      expect(mockDb.runAsync).toHaveBeenCalledWith(
        expect.stringContaining('UPDATE contacts'),
        expect.any(Array),
      );
    });

    it('inserts when no existing contact', async () => {
      mockDb.getFirstAsync.mockResolvedValue(null);
      mockDb.runAsync.mockResolvedValue(undefined);

      const id = await contactRepo.upsertContact(baseContact);

      expect(id).toMatch(/^local_/);
      expect(mockDb.runAsync).toHaveBeenCalledWith(
        expect.stringContaining('INSERT INTO contacts'),
        expect.any(Array),
      );
    });
  });

  describe('getContacts', () => {
    it('returns all contacts', async () => {
      mockDb.getAllAsync.mockResolvedValue([{ id: 'c1', name: 'Alice' }]);
      const result = await contactRepo.getContacts();
      expect(result).toHaveLength(1);
      expect(mockDb.getAllAsync).toHaveBeenCalledWith(expect.stringContaining('ORDER BY name'));
    });
  });

  describe('getRegisteredContacts', () => {
    it('returns only registered contacts', async () => {
      mockDb.getAllAsync.mockResolvedValue([{ id: 'c1', is_registered: 1 }]);
      const result = await contactRepo.getRegisteredContacts();
      expect(result).toHaveLength(1);
      expect(mockDb.getAllAsync).toHaveBeenCalledWith(expect.stringContaining('is_registered = 1'));
    });
  });

  describe('searchContacts', () => {
    it('searches by name', async () => {
      mockDb.getAllAsync.mockResolvedValue([{ id: 'c1', name: 'Alice' }]);
      const result = await contactRepo.searchContacts('Ali');
      expect(result).toHaveLength(1);
      expect(mockDb.getAllAsync).toHaveBeenCalledWith(
        expect.stringContaining('LIKE'),
        ['%Ali%'],
      );
    });
  });

  describe('deleteContact', () => {
    it('deletes by server_id', async () => {
      mockDb.runAsync.mockResolvedValue(undefined);
      await contactRepo.deleteContact('srv1');
      expect(mockDb.runAsync).toHaveBeenCalledWith(
        expect.stringContaining('DELETE FROM contacts'),
        ['srv1'],
      );
    });
  });

  describe('getContactCount', () => {
    it('returns count', async () => {
      mockDb.getFirstAsync.mockResolvedValue({ count: 5 });
      const result = await contactRepo.getContactCount();
      expect(result).toBe(5);
    });

    it('returns 0 when null', async () => {
      mockDb.getFirstAsync.mockResolvedValue(null);
      const result = await contactRepo.getContactCount();
      expect(result).toBe(0);
    });
  });
});
