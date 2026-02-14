// @ts-nocheck
jest.mock('../client', () => ({
  __esModule: true,
  default: {
    get: jest.fn(),
    post: jest.fn(),
    put: jest.fn(),
    delete: jest.fn(),
  },
}));

import apiClient from '../client';
import { backupApi } from '../backup';

const mock = apiClient as jest.Mocked<typeof apiClient>;

beforeEach(() => jest.clearAllMocks());

describe('backupApi', () => {
  it('export calls get', async () => {
    mock.get.mockResolvedValue({ data: {} });
    await backupApi.export();
    expect(mock.get).toHaveBeenCalledWith('/backup/export');
  });

  it('import calls post', async () => {
    mock.post.mockResolvedValue({ data: {} });
    const bundle = { version: 1, userId: 'u1', createdAt: '', data: {} as any };
    await backupApi.import(bundle);
    expect(mock.post).toHaveBeenCalledWith('/backup/import', bundle);
  });

  it('log calls post', async () => {
    mock.post.mockResolvedValue({ data: {} });
    await backupApi.log({ sizeBytes: 100, platform: 'google_drive', status: 'completed' });
    expect(mock.post).toHaveBeenCalledWith('/backup/log', expect.objectContaining({ sizeBytes: 100 }));
  });

  it('getHistory calls get', async () => {
    mock.get.mockResolvedValue({ data: [] });
    await backupApi.getHistory();
    expect(mock.get).toHaveBeenCalledWith('/backup/history');
  });

  it('getLatest calls get', async () => {
    mock.get.mockResolvedValue({ data: null });
    await backupApi.getLatest();
    expect(mock.get).toHaveBeenCalledWith('/backup/latest');
  });
});
