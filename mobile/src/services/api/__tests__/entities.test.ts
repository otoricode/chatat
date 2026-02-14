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
import { entitiesApi } from '../entities';

const mock = apiClient as jest.Mocked<typeof apiClient>;

beforeEach(() => jest.clearAllMocks());

describe('entitiesApi', () => {
  it('create calls post', async () => {
    mock.post.mockResolvedValue({ data: {} });
    await entitiesApi.create({ name: 'E1', type: 'person' });
    expect(mock.post).toHaveBeenCalledWith('/entities', { name: 'E1', type: 'person' });
  });

  it('list calls get', async () => {
    mock.get.mockResolvedValue({ data: [] });
    await entitiesApi.list({ type: 'person' });
    expect(mock.get).toHaveBeenCalledWith('/entities', { params: { type: 'person' } });
  });

  it('getById calls get', async () => {
    mock.get.mockResolvedValue({ data: {} });
    await entitiesApi.getById('e1');
    expect(mock.get).toHaveBeenCalledWith('/entities/e1');
  });

  it('update calls put', async () => {
    mock.put.mockResolvedValue({ data: {} });
    await entitiesApi.update('e1', { name: 'Updated' });
    expect(mock.put).toHaveBeenCalledWith('/entities/e1', { name: 'Updated' });
  });

  it('delete calls delete', async () => {
    mock.delete.mockResolvedValue({});
    await entitiesApi.delete('e1');
    expect(mock.delete).toHaveBeenCalledWith('/entities/e1');
  });

  it('search calls get', async () => {
    mock.get.mockResolvedValue({ data: [] });
    await entitiesApi.search('query');
    expect(mock.get).toHaveBeenCalledWith('/entities/search', { params: { q: 'query' } });
  });

  it('listTypes calls get', async () => {
    mock.get.mockResolvedValue({ data: [] });
    await entitiesApi.listTypes();
    expect(mock.get).toHaveBeenCalledWith('/entities/types');
  });

  it('createFromContact calls post', async () => {
    mock.post.mockResolvedValue({ data: {} });
    await entitiesApi.createFromContact('u1');
    expect(mock.post).toHaveBeenCalledWith('/entities/from-contact', { contactUserId: 'u1' });
  });

  it('linkToDocument calls post', async () => {
    mock.post.mockResolvedValue({});
    await entitiesApi.linkToDocument('d1', 'e1');
    expect(mock.post).toHaveBeenCalledWith('/documents/d1/entities', { entityId: 'e1' });
  });

  it('unlinkFromDocument calls delete', async () => {
    mock.delete.mockResolvedValue({});
    await entitiesApi.unlinkFromDocument('d1', 'e1');
    expect(mock.delete).toHaveBeenCalledWith('/documents/d1/entities/e1');
  });

  it('listDocuments calls get', async () => {
    mock.get.mockResolvedValue({ data: [] });
    await entitiesApi.listDocuments('e1');
    expect(mock.get).toHaveBeenCalledWith('/entities/e1/documents');
  });

  it('getDocumentEntities calls get', async () => {
    mock.get.mockResolvedValue({ data: [] });
    await entitiesApi.getDocumentEntities('d1');
    expect(mock.get).toHaveBeenCalledWith('/documents/d1/entities');
  });
});
