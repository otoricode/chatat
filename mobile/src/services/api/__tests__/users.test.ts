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
import { usersApi } from '../users';

const mock = apiClient as jest.Mocked<typeof apiClient>;

beforeEach(() => jest.clearAllMocks());

describe('usersApi', () => {
  it('getMe calls get', async () => {
    mock.get.mockResolvedValue({ data: {} });
    await usersApi.getMe();
    expect(mock.get).toHaveBeenCalledWith('/users/me');
  });

  it('updateMe calls put', async () => {
    mock.put.mockResolvedValue({ data: {} });
    await usersApi.updateMe({ name: 'New' });
    expect(mock.put).toHaveBeenCalledWith('/users/me', { name: 'New' });
  });

  it('setupProfile calls post', async () => {
    mock.post.mockResolvedValue({ data: {} });
    await usersApi.setupProfile({ name: 'N', avatar: 'a.jpg' });
    expect(mock.post).toHaveBeenCalledWith('/users/me/setup', { name: 'N', avatar: 'a.jpg' });
  });

  it('deleteAccount calls delete', async () => {
    mock.delete.mockResolvedValue({});
    await usersApi.deleteAccount();
    expect(mock.delete).toHaveBeenCalledWith('/users/me');
  });

  it('getPrivacy calls get', async () => {
    mock.get.mockResolvedValue({ data: {} });
    await usersApi.getPrivacy();
    expect(mock.get).toHaveBeenCalledWith('/users/me/privacy');
  });

  it('updatePrivacy calls put', async () => {
    mock.put.mockResolvedValue({ data: {} });
    await usersApi.updatePrivacy({ readReceipts: false });
    expect(mock.put).toHaveBeenCalledWith('/users/me/privacy', { readReceipts: false });
  });
});
