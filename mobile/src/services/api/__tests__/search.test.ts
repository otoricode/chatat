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
import { searchApi } from '../search';

const mock = apiClient as jest.Mocked<typeof apiClient>;

beforeEach(() => jest.clearAllMocks());

describe('searchApi', () => {
  it('searchAll calls get', async () => {
    mock.get.mockResolvedValue({ data: {} });
    await searchApi.searchAll('hello');
    expect(mock.get).toHaveBeenCalledWith('/search', { params: { q: 'hello', limit: 3 } });
  });

  it('searchAll with custom limit', async () => {
    mock.get.mockResolvedValue({ data: {} });
    await searchApi.searchAll('hello', 10);
    expect(mock.get).toHaveBeenCalledWith('/search', { params: { q: 'hello', limit: 10 } });
  });

  it('searchMessages calls get', async () => {
    mock.get.mockResolvedValue({ data: [] });
    await searchApi.searchMessages('query');
    expect(mock.get).toHaveBeenCalledWith('/search/messages', { params: { q: 'query', offset: 0, limit: 20 } });
  });

  it('searchDocuments calls get', async () => {
    mock.get.mockResolvedValue({ data: [] });
    await searchApi.searchDocuments('doc');
    expect(mock.get).toHaveBeenCalledWith('/search/documents', { params: { q: 'doc', offset: 0, limit: 20 } });
  });

  it('searchContacts calls get', async () => {
    mock.get.mockResolvedValue({ data: [] });
    await searchApi.searchContacts('Alice');
    expect(mock.get).toHaveBeenCalledWith('/search/contacts', { params: { q: 'Alice' } });
  });

  it('searchEntities calls get', async () => {
    mock.get.mockResolvedValue({ data: [] });
    await searchApi.searchEntities('company');
    expect(mock.get).toHaveBeenCalledWith('/search/entities', { params: { q: 'company' } });
  });

  it('searchInChat calls get', async () => {
    mock.get.mockResolvedValue({ data: [] });
    await searchApi.searchInChat('c1', 'hello');
    expect(mock.get).toHaveBeenCalledWith('/chats/c1/search', { params: { q: 'hello', offset: 0, limit: 20 } });
  });
});
