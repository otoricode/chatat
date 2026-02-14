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
import { contactsApi } from '../contacts';

const mock = apiClient as jest.Mocked<typeof apiClient>;

beforeEach(() => jest.clearAllMocks());

describe('contactsApi', () => {
  it('sync calls post', async () => {
    mock.post.mockResolvedValue({ data: { matched: 5 } });
    await contactsApi.sync({ phoneHashes: ['h1'] });
    expect(mock.post).toHaveBeenCalledWith('/contacts/sync', { phoneHashes: ['h1'] });
  });

  it('list calls get', async () => {
    mock.get.mockResolvedValue({ data: [] });
    await contactsApi.list();
    expect(mock.get).toHaveBeenCalledWith('/contacts');
  });

  it('search calls get with query', async () => {
    mock.get.mockResolvedValue({ data: [] });
    await contactsApi.search('+628');
    expect(mock.get).toHaveBeenCalledWith(expect.stringContaining('/contacts/search'));
  });

  it('getProfile calls get', async () => {
    mock.get.mockResolvedValue({ data: {} });
    await contactsApi.getProfile('u1');
    expect(mock.get).toHaveBeenCalledWith('/contacts/u1');
  });
});
