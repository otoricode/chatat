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
import { chatsApi } from '../chats';

const mock = apiClient as jest.Mocked<typeof apiClient>;

beforeEach(() => jest.clearAllMocks());

describe('chatsApi', () => {
  it('list calls get /chats', async () => {
    mock.get.mockResolvedValue({ data: [] });
    await chatsApi.list();
    expect(mock.get).toHaveBeenCalledWith('/chats');
  });

  it('create calls post /chats', async () => {
    mock.post.mockResolvedValue({ data: {} });
    await chatsApi.create('c1');
    expect(mock.post).toHaveBeenCalledWith('/chats', { contactId: 'c1' });
  });

  it('createGroup calls post /chats', async () => {
    mock.post.mockResolvedValue({ data: {} });
    await chatsApi.createGroup({ type: 'group', name: 'G', icon: '', memberIds: [] });
    expect(mock.post).toHaveBeenCalledWith('/chats', expect.objectContaining({ type: 'group' }));
  });

  it('getById calls get /chats/:id', async () => {
    mock.get.mockResolvedValue({ data: {} });
    await chatsApi.getById('c1');
    expect(mock.get).toHaveBeenCalledWith('/chats/c1');
  });

  it('pinChat calls put', async () => {
    mock.put.mockResolvedValue({});
    await chatsApi.pinChat('c1');
    expect(mock.put).toHaveBeenCalledWith('/chats/c1/pin');
  });

  it('unpinChat calls delete', async () => {
    mock.delete.mockResolvedValue({});
    await chatsApi.unpinChat('c1');
    expect(mock.delete).toHaveBeenCalledWith('/chats/c1/pin');
  });

  it('markAsRead calls post', async () => {
    mock.post.mockResolvedValue({});
    await chatsApi.markAsRead('c1');
    expect(mock.post).toHaveBeenCalledWith('/chats/c1/read');
  });

  it('getMessages without params', async () => {
    mock.get.mockResolvedValue({ data: [] });
    await chatsApi.getMessages('c1');
    expect(mock.get).toHaveBeenCalledWith('/chats/c1/messages');
  });

  it('getMessages with cursor and limit', async () => {
    mock.get.mockResolvedValue({ data: [] });
    await chatsApi.getMessages('c1', 'cur1', 10);
    expect(mock.get).toHaveBeenCalledWith(
      expect.stringContaining('cursor=cur1'),
    );
  });

  it('sendMessage calls post', async () => {
    mock.post.mockResolvedValue({ data: {} });
    await chatsApi.sendMessage('c1', { content: 'hi' });
    expect(mock.post).toHaveBeenCalledWith('/chats/c1/messages', { content: 'hi' });
  });

  it('deleteMessage calls delete', async () => {
    mock.delete.mockResolvedValue({});
    await chatsApi.deleteMessage('c1', 'm1', true);
    expect(mock.delete).toHaveBeenCalledWith('/chats/c1/messages/m1?forAll=true');
  });

  it('searchMessages calls get', async () => {
    mock.get.mockResolvedValue({ data: [] });
    await chatsApi.searchMessages('c1', 'hello');
    expect(mock.get).toHaveBeenCalledWith(expect.stringContaining('search'));
  });

  it('addMember calls post', async () => {
    mock.post.mockResolvedValue({});
    await chatsApi.addMember('c1', 'u1');
    expect(mock.post).toHaveBeenCalledWith('/chats/c1/members', { userId: 'u1' });
  });

  it('removeMember calls delete', async () => {
    mock.delete.mockResolvedValue({});
    await chatsApi.removeMember('c1', 'u1');
    expect(mock.delete).toHaveBeenCalledWith('/chats/c1/members/u1');
  });

  it('leaveGroup calls post', async () => {
    mock.post.mockResolvedValue({});
    await chatsApi.leaveGroup('c1');
    expect(mock.post).toHaveBeenCalledWith('/chats/c1/leave');
  });

  it('updateGroup calls put', async () => {
    mock.put.mockResolvedValue({ data: {} });
    await chatsApi.updateGroup('c1', { name: 'New Name' });
    expect(mock.put).toHaveBeenCalledWith('/chats/c1', { name: 'New Name' });
  });

  it('deleteGroup calls delete', async () => {
    mock.delete.mockResolvedValue({});
    await chatsApi.deleteGroup('c1');
    expect(mock.delete).toHaveBeenCalledWith('/chats/c1');
  });

  it('getGroupInfo calls get', async () => {
    mock.get.mockResolvedValue({ data: {} });
    await chatsApi.getGroupInfo('c1');
    expect(mock.get).toHaveBeenCalledWith('/chats/c1/info');
  });

  it('promoteToAdmin calls put', async () => {
    mock.put.mockResolvedValue({});
    await chatsApi.promoteToAdmin('c1', 'u1');
    expect(mock.put).toHaveBeenCalledWith('/chats/c1/members/u1/admin');
  });

  it('forwardMessage calls post', async () => {
    mock.post.mockResolvedValue({ data: {} });
    await chatsApi.forwardMessage('c1', 'm1', { targetChatId: 'c2' });
    expect(mock.post).toHaveBeenCalledWith('/chats/c1/messages/m1/forward', { targetChatId: 'c2' });
  });
});
