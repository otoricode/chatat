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
import { documentsApi } from '../documents';

const mock = apiClient as jest.Mocked<typeof apiClient>;

beforeEach(() => jest.clearAllMocks());

describe('documentsApi', () => {
  it('create calls post', async () => {
    mock.post.mockResolvedValue({ data: {} });
    await documentsApi.create({ title: 'Doc' });
    expect(mock.post).toHaveBeenCalledWith('/documents', { title: 'Doc' });
  });

  it('getById calls get', async () => {
    mock.get.mockResolvedValue({ data: {} });
    await documentsApi.getById('d1');
    expect(mock.get).toHaveBeenCalledWith('/documents/d1');
  });

  it('list calls get with params', async () => {
    mock.get.mockResolvedValue({ data: [] });
    await documentsApi.list('cur1', 10);
    expect(mock.get).toHaveBeenCalledWith('/documents', { params: { cursor: 'cur1', limit: 10 } });
  });

  it('listByChat calls get', async () => {
    mock.get.mockResolvedValue({ data: [] });
    await documentsApi.listByChat('c1');
    expect(mock.get).toHaveBeenCalledWith('/chats/c1/documents');
  });

  it('listByTopic calls get', async () => {
    mock.get.mockResolvedValue({ data: [] });
    await documentsApi.listByTopic('t1');
    expect(mock.get).toHaveBeenCalledWith('/topics/t1/documents');
  });

  it('update calls put', async () => {
    mock.put.mockResolvedValue({ data: {} });
    await documentsApi.update('d1', { title: 'Updated' });
    expect(mock.put).toHaveBeenCalledWith('/documents/d1', { title: 'Updated' });
  });

  it('delete calls delete', async () => {
    mock.delete.mockResolvedValue({});
    await documentsApi.delete('d1');
    expect(mock.delete).toHaveBeenCalledWith('/documents/d1');
  });

  it('duplicate calls post', async () => {
    mock.post.mockResolvedValue({ data: {} });
    await documentsApi.duplicate('d1');
    expect(mock.post).toHaveBeenCalledWith('/documents/d1/duplicate');
  });

  it('addBlock calls post', async () => {
    mock.post.mockResolvedValue({ data: {} });
    await documentsApi.addBlock('d1', { type: 'paragraph' as any, sortOrder: 0 });
    expect(mock.post).toHaveBeenCalledWith('/documents/d1/blocks', expect.any(Object));
  });

  it('updateBlock calls put', async () => {
    mock.put.mockResolvedValue({ data: {} });
    await documentsApi.updateBlock('d1', 'b1', { content: 'new' } as any);
    expect(mock.put).toHaveBeenCalledWith('/documents/d1/blocks/b1', { content: 'new' });
  });

  it('deleteBlock calls delete', async () => {
    mock.delete.mockResolvedValue({});
    await documentsApi.deleteBlock('d1', 'b1');
    expect(mock.delete).toHaveBeenCalledWith('/documents/d1/blocks/b1');
  });

  it('reorderBlocks calls put', async () => {
    mock.put.mockResolvedValue({});
    await documentsApi.reorderBlocks('d1', ['b1', 'b2']);
    expect(mock.put).toHaveBeenCalledWith('/documents/d1/blocks/reorder', { blockIds: ['b1', 'b2'] });
  });

  it('addCollaborator calls post', async () => {
    mock.post.mockResolvedValue({});
    await documentsApi.addCollaborator('d1', 'u1', 'editor' as any);
    expect(mock.post).toHaveBeenCalledWith('/documents/d1/collaborators', { userId: 'u1', role: 'editor' });
  });

  it('removeCollaborator calls delete', async () => {
    mock.delete.mockResolvedValue({});
    await documentsApi.removeCollaborator('d1', 'u1');
    expect(mock.delete).toHaveBeenCalledWith('/documents/d1/collaborators/u1');
  });

  it('getHistory calls get', async () => {
    mock.get.mockResolvedValue({ data: [] });
    await documentsApi.getHistory('d1');
    expect(mock.get).toHaveBeenCalledWith('/documents/d1/history');
  });

  it('getTemplates calls get', async () => {
    mock.get.mockResolvedValue({ data: [] });
    await documentsApi.getTemplates();
    expect(mock.get).toHaveBeenCalledWith('/templates');
  });

  it('lock calls post', async () => {
    mock.post.mockResolvedValue({});
    await documentsApi.lock('d1', 'manual');
    expect(mock.post).toHaveBeenCalledWith('/documents/d1/lock', { mode: 'manual' });
  });

  it('unlock calls post', async () => {
    mock.post.mockResolvedValue({});
    await documentsApi.unlock('d1');
    expect(mock.post).toHaveBeenCalledWith('/documents/d1/unlock');
  });

  it('sign calls post', async () => {
    mock.post.mockResolvedValue({});
    await documentsApi.sign('d1', 'John');
    expect(mock.post).toHaveBeenCalledWith('/documents/d1/sign', { name: 'John' });
  });

  it('listSigners calls get', async () => {
    mock.get.mockResolvedValue({ data: [] });
    await documentsApi.listSigners('d1');
    expect(mock.get).toHaveBeenCalledWith('/documents/d1/signers');
  });

  it('addSigner calls post', async () => {
    mock.post.mockResolvedValue({});
    await documentsApi.addSigner('d1', 'u1');
    expect(mock.post).toHaveBeenCalledWith('/documents/d1/signers', { userId: 'u1' });
  });

  it('removeSigner calls delete', async () => {
    mock.delete.mockResolvedValue({});
    await documentsApi.removeSigner('d1', 'u1');
    expect(mock.delete).toHaveBeenCalledWith('/documents/d1/signers/u1');
  });

  it('updateCollaboratorRole calls put', async () => {
    mock.put.mockResolvedValue({});
    await documentsApi.updateCollaboratorRole('d1', 'u1', 'viewer' as any);
    expect(mock.put).toHaveBeenCalledWith('/documents/d1/collaborators/u1', { role: 'viewer' });
  });

  it('addTag calls post', async () => {
    mock.post.mockResolvedValue({});
    await documentsApi.addTag('d1', 'important');
    expect(mock.post).toHaveBeenCalledWith('/documents/d1/tags', { tag: 'important' });
  });

  it('removeTag calls delete', async () => {
    mock.delete.mockResolvedValue({});
    await documentsApi.removeTag('d1', 'important');
    expect(mock.delete).toHaveBeenCalledWith('/documents/d1/tags/important');
  });

  it('batchBlocks calls post', async () => {
    mock.post.mockResolvedValue({});
    await documentsApi.batchBlocks('d1', [{ action: 'add' }]);
    expect(mock.post).toHaveBeenCalledWith('/documents/d1/blocks/batch', { operations: [{ action: 'add' }] });
  });
});
