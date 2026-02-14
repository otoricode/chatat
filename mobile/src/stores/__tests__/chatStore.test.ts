// @ts-nocheck
jest.mock('@/services/api/chats', () => ({
  chatsApi: {
    list: jest.fn(),
    pinChat: jest.fn(),
    unpinChat: jest.fn(),
    markAsRead: jest.fn(),
  },
}));

import { useChatStore } from '../chatStore';
import { chatsApi } from '@/services/api/chats';

const mockChatsApi = chatsApi as jest.Mocked<typeof chatsApi>;

const makeChatItem = (id: string, name: string, unread = 0) => ({
  chat: { id, type: 'personal' as const, name, icon: null, pinnedAt: null, createdAt: '2024-01-01' },
  lastMessage: null,
  unreadCount: unread,
  otherUser: null,
  isOnline: false,
});

beforeEach(() => {
  useChatStore.setState({
    chats: [],
    isLoading: false,
    error: null,
  });
  jest.clearAllMocks();
});

describe('chatStore', () => {
  it('starts with empty chats', () => {
    const s = useChatStore.getState();
    expect(s.chats).toEqual([]);
    expect(s.isLoading).toBe(false);
    expect(s.error).toBeNull();
  });

  it('fetchChats success', async () => {
    const items = [makeChatItem('1', 'Alice'), makeChatItem('2', 'Bob')];
    mockChatsApi.list.mockResolvedValue({ data: { data: items } } as any);

    await useChatStore.getState().fetchChats();

    const s = useChatStore.getState();
    expect(s.chats).toHaveLength(2);
    expect(s.isLoading).toBe(false);
    expect(s.error).toBeNull();
  });

  it('fetchChats handles null data', async () => {
    mockChatsApi.list.mockResolvedValue({ data: { data: null } } as any);

    await useChatStore.getState().fetchChats();

    expect(useChatStore.getState().chats).toEqual([]);
  });

  it('fetchChats error with Error instance', async () => {
    mockChatsApi.list.mockRejectedValue(new Error('Network error'));

    await useChatStore.getState().fetchChats();

    const s = useChatStore.getState();
    expect(s.error).toBe('Network error');
    expect(s.isLoading).toBe(false);
  });

  it('fetchChats error with non-Error', async () => {
    mockChatsApi.list.mockRejectedValue('some string');

    await useChatStore.getState().fetchChats();

    expect(useChatStore.getState().error).toBe('Failed to load chats');
  });

  it('updateLastMessage updates correct chat', () => {
    const items = [makeChatItem('1', 'Alice', 3), makeChatItem('2', 'Bob', 0)];
    useChatStore.setState({ chats: items as any });

    const msg = { id: 'm1', content: 'Hi', type: 'text' } as any;
    useChatStore.getState().updateLastMessage('1', msg);

    const s = useChatStore.getState();
    expect((s.chats[0] as any).lastMessage).toBe(msg);
    expect(s.chats[0].unreadCount).toBe(4); // incremented
    expect(s.chats[1].unreadCount).toBe(0); // unchanged
  });

  it('updateUnreadCount updates correct chat', () => {
    const items = [makeChatItem('1', 'Alice', 3), makeChatItem('2', 'Bob', 5)];
    useChatStore.setState({ chats: items as any });

    useChatStore.getState().updateUnreadCount('2', 0);

    expect(useChatStore.getState().chats[1].unreadCount).toBe(0);
    expect(useChatStore.getState().chats[0].unreadCount).toBe(3);
  });

  it('pinChat success calls API and refetches', async () => {
    mockChatsApi.pinChat.mockResolvedValue({} as any);
    mockChatsApi.list.mockResolvedValue({ data: { data: [] } } as any);

    await useChatStore.getState().pinChat('1');

    expect(mockChatsApi.pinChat).toHaveBeenCalledWith('1');
    expect(mockChatsApi.list).toHaveBeenCalled();
  });

  it('pinChat error with Error instance', async () => {
    mockChatsApi.pinChat.mockRejectedValue(new Error('Pin failed'));

    await useChatStore.getState().pinChat('1');

    expect(useChatStore.getState().error).toBe('Pin failed');
  });

  it('pinChat error with non-Error', async () => {
    mockChatsApi.pinChat.mockRejectedValue(42);

    await useChatStore.getState().pinChat('1');

    expect(useChatStore.getState().error).toBe('Failed to pin chat');
  });

  it('unpinChat success', async () => {
    mockChatsApi.unpinChat.mockResolvedValue({} as any);
    mockChatsApi.list.mockResolvedValue({ data: { data: [] } } as any);

    await useChatStore.getState().unpinChat('1');

    expect(mockChatsApi.unpinChat).toHaveBeenCalledWith('1');
  });

  it('unpinChat error with Error instance', async () => {
    mockChatsApi.unpinChat.mockRejectedValue(new Error('Unpin failed'));

    await useChatStore.getState().unpinChat('1');

    expect(useChatStore.getState().error).toBe('Unpin failed');
  });

  it('unpinChat error with non-Error', async () => {
    mockChatsApi.unpinChat.mockRejectedValue('err');

    await useChatStore.getState().unpinChat('1');

    expect(useChatStore.getState().error).toBe('Failed to unpin chat');
  });

  it('markAsRead success', async () => {
    const items = [makeChatItem('1', 'Alice', 5)];
    useChatStore.setState({ chats: items as any });
    mockChatsApi.markAsRead.mockResolvedValue({} as any);

    await useChatStore.getState().markAsRead('1');

    expect(useChatStore.getState().chats[0].unreadCount).toBe(0);
  });

  it('markAsRead error is silent', async () => {
    mockChatsApi.markAsRead.mockRejectedValue(new Error('fail'));

    await useChatStore.getState().markAsRead('1');

    // No error set
    expect(useChatStore.getState().error).toBeNull();
  });

  it('clearError resets error', () => {
    useChatStore.setState({ error: 'some error' });

    useChatStore.getState().clearError();

    expect(useChatStore.getState().error).toBeNull();
  });

  it('updateChatOnlineStatus updates matching user', () => {
    const items = [
      {
        ...makeChatItem('1', 'Alice'),
        otherUser: { id: 'u1', name: 'Alice', phone: '+62', avatar: '', lastSeen: '' },
        isOnline: false,
      },
    ];
    useChatStore.setState({ chats: items as any });

    useChatStore.getState().updateChatOnlineStatus('u1', true, '2024-01-01T12:00:00Z');

    const s = useChatStore.getState();
    expect((s.chats[0] as any).isOnline).toBe(true);
    expect((s.chats[0] as any).otherUser.lastSeen).toBe('2024-01-01T12:00:00Z');
  });

  it('updateChatOnlineStatus ignores non-matching user', () => {
    const items = [
      {
        ...makeChatItem('1', 'Alice'),
        otherUser: { id: 'u1', name: 'Alice', phone: '+62', avatar: '', lastSeen: '' },
        isOnline: false,
      },
    ];
    useChatStore.setState({ chats: items as any });

    useChatStore.getState().updateChatOnlineStatus('u999', true, '2024-01-01');

    expect((useChatStore.getState().chats[0] as any).isOnline).toBe(false);
  });
});
