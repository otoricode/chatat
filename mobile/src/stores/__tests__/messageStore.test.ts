// @ts-nocheck
jest.mock('@/services/api/chats', () => ({
  chatsApi: {
    getMessages: jest.fn(),
  },
}));

import { useMessageStore } from '../messageStore';
import { chatsApi } from '@/services/api/chats';

const mockChatsApi = chatsApi as jest.Mocked<typeof chatsApi>;

const makeMsg = (id: string, content: string) => ({
  id,
  senderId: 's1',
  content,
  type: 'text',
  createdAt: '2024-01-01T00:00:00Z',
  isDeleted: false,
  deletedForAll: false,
});

beforeEach(() => {
  useMessageStore.setState({
    messages: {},
    isLoading: false,
    hasMore: {},
    cursor: {},
    error: null,
  });
  jest.clearAllMocks();
});

describe('messageStore', () => {
  it('starts empty', () => {
    const s = useMessageStore.getState();
    expect(s.messages).toEqual({});
    expect(s.isLoading).toBe(false);
    expect(s.error).toBeNull();
  });

  it('fetchMessages success', async () => {
    const msgs = [makeMsg('m1', 'Hello'), makeMsg('m2', 'World')];
    mockChatsApi.getMessages.mockResolvedValue({
      data: { data: msgs, meta: { hasMore: true, cursor: 'cur1' } },
    } as any);

    await useMessageStore.getState().fetchMessages('chat1');

    const s = useMessageStore.getState();
    expect(s.messages['chat1']).toHaveLength(2);
    expect(s.hasMore['chat1']).toBe(true);
    expect(s.cursor['chat1']).toBe('cur1');
    expect(s.isLoading).toBe(false);
  });

  it('fetchMessages handles null data', async () => {
    mockChatsApi.getMessages.mockResolvedValue({
      data: { data: null, meta: { hasMore: false, cursor: '' } },
    } as any);

    await useMessageStore.getState().fetchMessages('chat1');

    expect(useMessageStore.getState().messages['chat1']).toEqual([]);
  });

  it('fetchMessages error with Error', async () => {
    mockChatsApi.getMessages.mockRejectedValue(new Error('Network error'));

    await useMessageStore.getState().fetchMessages('chat1');

    expect(useMessageStore.getState().error).toBe('Network error');
    expect(useMessageStore.getState().isLoading).toBe(false);
  });

  it('fetchMessages error with non-Error', async () => {
    mockChatsApi.getMessages.mockRejectedValue('fail');

    await useMessageStore.getState().fetchMessages('chat1');

    expect(useMessageStore.getState().error).toBe('Failed to load messages');
  });

  it('fetchMore appends messages', async () => {
    useMessageStore.setState({
      messages: { chat1: [makeMsg('m1', 'First')] as any },
      hasMore: { chat1: true },
      cursor: { chat1: 'cur1' },
    });

    const newMsgs = [makeMsg('m2', 'Second')];
    mockChatsApi.getMessages.mockResolvedValue({
      data: { data: newMsgs, meta: { hasMore: false, cursor: 'cur2' } },
    } as any);

    await useMessageStore.getState().fetchMore('chat1');

    expect(useMessageStore.getState().messages['chat1']).toHaveLength(2);
    expect(useMessageStore.getState().hasMore['chat1']).toBe(false);
  });

  it('fetchMore skips when no hasMore', async () => {
    useMessageStore.setState({
      messages: { chat1: [] },
      hasMore: { chat1: false },
      cursor: { chat1: 'cur1' },
    });

    await useMessageStore.getState().fetchMore('chat1');

    expect(mockChatsApi.getMessages).not.toHaveBeenCalled();
  });

  it('fetchMore skips when isLoading', async () => {
    useMessageStore.setState({
      messages: { chat1: [] },
      hasMore: { chat1: true },
      cursor: { chat1: 'cur1' },
      isLoading: true,
    });

    await useMessageStore.getState().fetchMore('chat1');

    expect(mockChatsApi.getMessages).not.toHaveBeenCalled();
  });

  it('fetchMore skips when no cursor', async () => {
    useMessageStore.setState({
      messages: { chat1: [] },
      hasMore: { chat1: true },
      cursor: {},
    });

    await useMessageStore.getState().fetchMore('chat1');

    expect(mockChatsApi.getMessages).not.toHaveBeenCalled();
  });

  it('fetchMore error with Error', async () => {
    useMessageStore.setState({
      messages: { chat1: [] },
      hasMore: { chat1: true },
      cursor: { chat1: 'cur1' },
    });

    mockChatsApi.getMessages.mockRejectedValue(new Error('Fail'));

    await useMessageStore.getState().fetchMore('chat1');

    expect(useMessageStore.getState().error).toBe('Fail');
  });

  it('fetchMore error with non-Error', async () => {
    useMessageStore.setState({
      messages: { chat1: [] },
      hasMore: { chat1: true },
      cursor: { chat1: 'cur1' },
    });

    mockChatsApi.getMessages.mockRejectedValue(42);

    await useMessageStore.getState().fetchMore('chat1');

    expect(useMessageStore.getState().error).toBe('Failed to load more messages');
  });

  it('addMessage prepends to chat', () => {
    useMessageStore.setState({ messages: { chat1: [makeMsg('m1', 'First')] as any } });

    const newMsg = makeMsg('m2', 'New') as any;
    useMessageStore.getState().addMessage('chat1', newMsg);

    const msgs = useMessageStore.getState().messages['chat1'];
    expect(msgs).toHaveLength(2);
    expect(msgs[0].id).toBe('m2'); // prepended
  });

  it('addMessage creates array for new chat', () => {
    useMessageStore.getState().addMessage('chat2', makeMsg('m1', 'Hi') as any);

    expect(useMessageStore.getState().messages['chat2']).toHaveLength(1);
  });

  it('deleteMessage marks as deleted', () => {
    useMessageStore.setState({ messages: { chat1: [makeMsg('m1', 'Hello')] as any } });

    useMessageStore.getState().deleteMessage('chat1', 'm1');

    const msg = useMessageStore.getState().messages['chat1'][0] as any;
    expect(msg.isDeleted).toBe(true);
    expect(msg.deletedForAll).toBe(true);
    expect(msg.content).toBe('');
  });

  it('clearMessages removes chat data', () => {
    useMessageStore.setState({
      messages: { chat1: [makeMsg('m1', 'Hello')] as any },
      hasMore: { chat1: true },
      cursor: { chat1: 'cur1' },
    });

    useMessageStore.getState().clearMessages('chat1');

    const s = useMessageStore.getState();
    expect(s.messages['chat1']).toBeUndefined();
    expect(s.hasMore['chat1']).toBeUndefined();
    expect(s.cursor['chat1']).toBeUndefined();
  });

  it('clearError resets error', () => {
    useMessageStore.setState({ error: 'some error' });

    useMessageStore.getState().clearError();

    expect(useMessageStore.getState().error).toBeNull();
  });
});
