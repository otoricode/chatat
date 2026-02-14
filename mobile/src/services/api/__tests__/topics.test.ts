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
import { topicsApi } from '../topics';

const mock = apiClient as jest.Mocked<typeof apiClient>;

beforeEach(() => jest.clearAllMocks());

describe('topicsApi', () => {
  it('list calls get /topics', async () => {
    mock.get.mockResolvedValue({ data: [] });
    await topicsApi.list();
    expect(mock.get).toHaveBeenCalledWith('/topics');
  });

  it('listByChat calls get', async () => {
    mock.get.mockResolvedValue({ data: [] });
    await topicsApi.listByChat('c1');
    expect(mock.get).toHaveBeenCalledWith('/chats/c1/topics');
  });

  it('create calls post', async () => {
    mock.post.mockResolvedValue({ data: {} });
    await topicsApi.create({ name: 'T', icon: '', parentId: 'c1' });
    expect(mock.post).toHaveBeenCalledWith('/topics', expect.objectContaining({ name: 'T' }));
  });

  it('getById calls get', async () => {
    mock.get.mockResolvedValue({ data: {} });
    await topicsApi.getById('t1');
    expect(mock.get).toHaveBeenCalledWith('/topics/t1');
  });

  it('update calls put', async () => {
    mock.put.mockResolvedValue({ data: {} });
    await topicsApi.update('t1', { name: 'Updated' });
    expect(mock.put).toHaveBeenCalledWith('/topics/t1', { name: 'Updated' });
  });

  it('delete calls delete', async () => {
    mock.delete.mockResolvedValue({});
    await topicsApi.delete('t1');
    expect(mock.delete).toHaveBeenCalledWith('/topics/t1');
  });

  it('getMessages without params', async () => {
    mock.get.mockResolvedValue({ data: [] });
    await topicsApi.getMessages('t1');
    expect(mock.get).toHaveBeenCalledWith('/topics/t1/messages');
  });

  it('getMessages with cursor and limit', async () => {
    mock.get.mockResolvedValue({ data: [] });
    await topicsApi.getMessages('t1', 'cur1', 5);
    expect(mock.get).toHaveBeenCalledWith(expect.stringContaining('cursor=cur1'));
  });

  it('sendMessage calls post', async () => {
    mock.post.mockResolvedValue({ data: {} });
    await topicsApi.sendMessage('t1', { content: 'hello' });
    expect(mock.post).toHaveBeenCalledWith('/topics/t1/messages', { content: 'hello' });
  });

  it('deleteMessage calls delete', async () => {
    mock.delete.mockResolvedValue({});
    await topicsApi.deleteMessage('t1', 'm1', true);
    expect(mock.delete).toHaveBeenCalledWith('/topics/t1/messages/m1?forAll=true');
  });

  it('addMember calls post', async () => {
    mock.post.mockResolvedValue({});
    await topicsApi.addMember('t1', 'u1');
    expect(mock.post).toHaveBeenCalledWith('/topics/t1/members', { userId: 'u1' });
  });

  it('removeMember calls delete', async () => {
    mock.delete.mockResolvedValue({});
    await topicsApi.removeMember('t1', 'u1');
    expect(mock.delete).toHaveBeenCalledWith('/topics/t1/members/u1');
  });
});
