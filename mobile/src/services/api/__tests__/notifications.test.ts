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
import { notificationsApi } from '../notifications';

const mock = apiClient as jest.Mocked<typeof apiClient>;

beforeEach(() => jest.clearAllMocks());

describe('notificationsApi', () => {
  it('registerDevice calls post', async () => {
    mock.post.mockResolvedValue({});
    await notificationsApi.registerDevice('tok123', 'ios');
    expect(mock.post).toHaveBeenCalledWith('/notifications/devices', { token: 'tok123', platform: 'ios' });
  });

  it('unregisterDevice calls delete', async () => {
    mock.delete.mockResolvedValue({});
    await notificationsApi.unregisterDevice('tok123');
    expect(mock.delete).toHaveBeenCalledWith('/notifications/devices', { data: { token: 'tok123' } });
  });
});
