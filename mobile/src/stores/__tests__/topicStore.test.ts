// @ts-nocheck
jest.mock('@/services/api/topics', () => ({
  topicsApi: {
    listByChat: jest.fn(),
    getMessages: jest.fn(),
  },
}));

import { useTopicStore } from '../topicStore';
import { topicsApi } from '@/services/api/topics';

const mockTopicsApi = topicsApi as jest.Mocked<typeof topicsApi>;

const makeTopic = (id: string, name: string) => ({
  topic: { id, name, chatId: 'c1', icon: '', createdAt: '2024-01-01' },
  messageCount: 0,
});

const makeTopicMsg = (id: string, content: string) => ({
  id,
  content,
  senderId: 's1',
  type: 'text',
  createdAt: '2024-01-01',
  isDeleted: false,
});

beforeEach(() => {
  useTopicStore.setState({
    topicsByChat: {},
    messagesByTopic: {},
    isLoading: false,
    error: null,
  });
  jest.clearAllMocks();
});

describe('topicStore', () => {
  it('starts empty', () => {
    const s = useTopicStore.getState();
    expect(s.topicsByChat).toEqual({});
    expect(s.messagesByTopic).toEqual({});
    expect(s.isLoading).toBe(false);
  });

  it('fetchTopics success', async () => {
    const topics = [makeTopic('t1', 'General'), makeTopic('t2', 'Random')];
    mockTopicsApi.listByChat.mockResolvedValue({ data: { data: topics } } as any);

    await useTopicStore.getState().fetchTopics('c1');

    const s = useTopicStore.getState();
    expect(s.topicsByChat['c1']).toHaveLength(2);
    expect(s.isLoading).toBe(false);
  });

  it('fetchTopics handles null data', async () => {
    mockTopicsApi.listByChat.mockResolvedValue({ data: { data: null } } as any);

    await useTopicStore.getState().fetchTopics('c1');

    expect(useTopicStore.getState().topicsByChat['c1']).toEqual([]);
  });

  it('fetchTopics error with Error', async () => {
    mockTopicsApi.listByChat.mockRejectedValue(new Error('Network error'));

    await useTopicStore.getState().fetchTopics('c1');

    expect(useTopicStore.getState().error).toBe('Network error');
    expect(useTopicStore.getState().isLoading).toBe(false);
  });

  it('fetchTopics error with non-Error', async () => {
    mockTopicsApi.listByChat.mockRejectedValue('fail');

    await useTopicStore.getState().fetchTopics('c1');

    expect(useTopicStore.getState().error).toBe('Failed to load topics');
  });

  it('addTopic prepends to chat', () => {
    useTopicStore.setState({ topicsByChat: { c1: [makeTopic('t1', 'Existing')] as any } });

    useTopicStore.getState().addTopic('c1', makeTopic('t2', 'New') as any);

    const topics = useTopicStore.getState().topicsByChat['c1'];
    expect(topics).toHaveLength(2);
    expect((topics[0] as any).topic.id).toBe('t2'); // prepended
  });

  it('addTopic creates array for new chat', () => {
    useTopicStore.getState().addTopic('c2', makeTopic('t1', 'First') as any);

    expect(useTopicStore.getState().topicsByChat['c2']).toHaveLength(1);
  });

  it('removeTopic removes from chat', () => {
    useTopicStore.setState({
      topicsByChat: { c1: [makeTopic('t1', 'A'), makeTopic('t2', 'B')] as any },
    });

    useTopicStore.getState().removeTopic('c1', 't1');

    const topics = useTopicStore.getState().topicsByChat['c1'];
    expect(topics).toHaveLength(1);
    expect((topics[0] as any).topic.id).toBe('t2');
  });

  it('fetchMessages success', async () => {
    const msgs = [makeTopicMsg('m1', 'Hello'), makeTopicMsg('m2', 'World')];
    mockTopicsApi.getMessages.mockResolvedValue({ data: { data: msgs } } as any);

    await useTopicStore.getState().fetchMessages('t1');

    expect(useTopicStore.getState().messagesByTopic['t1']).toHaveLength(2);
    expect(useTopicStore.getState().isLoading).toBe(false);
  });

  it('fetchMessages handles null data', async () => {
    mockTopicsApi.getMessages.mockResolvedValue({ data: { data: null } } as any);

    await useTopicStore.getState().fetchMessages('t1');

    expect(useTopicStore.getState().messagesByTopic['t1']).toEqual([]);
  });

  it('fetchMessages error with Error', async () => {
    mockTopicsApi.getMessages.mockRejectedValue(new Error('Fail'));

    await useTopicStore.getState().fetchMessages('t1');

    expect(useTopicStore.getState().error).toBe('Fail');
  });

  it('fetchMessages error with non-Error', async () => {
    mockTopicsApi.getMessages.mockRejectedValue(42);

    await useTopicStore.getState().fetchMessages('t1');

    expect(useTopicStore.getState().error).toBe('Failed to load messages');
  });

  it('addMessage prepends to topic', () => {
    useTopicStore.setState({
      messagesByTopic: { t1: [makeTopicMsg('m1', 'First')] as any },
    });

    useTopicStore.getState().addMessage('t1', makeTopicMsg('m2', 'New') as any);

    const msgs = useTopicStore.getState().messagesByTopic['t1'];
    expect(msgs).toHaveLength(2);
    expect(msgs[0].id).toBe('m2');
  });

  it('addMessage ignores duplicate', () => {
    useTopicStore.setState({
      messagesByTopic: { t1: [makeTopicMsg('m1', 'First')] as any },
    });

    useTopicStore.getState().addMessage('t1', makeTopicMsg('m1', 'Dup') as any);

    expect(useTopicStore.getState().messagesByTopic['t1']).toHaveLength(1);
  });

  it('addMessage creates array for new topic', () => {
    useTopicStore.getState().addMessage('t2', makeTopicMsg('m1', 'Hi') as any);

    expect(useTopicStore.getState().messagesByTopic['t2']).toHaveLength(1);
  });

  it('deleteMessage marks as deleted', () => {
    useTopicStore.setState({
      messagesByTopic: { t1: [makeTopicMsg('m1', 'Hello')] as any },
    });

    useTopicStore.getState().deleteMessage('t1', 'm1');

    const msg = useTopicStore.getState().messagesByTopic['t1'][0];
    expect((msg as any).isDeleted).toBe(true);
    expect(msg.content).toBe('');
  });

  it('clearError resets error', () => {
    useTopicStore.setState({ error: 'some error' });

    useTopicStore.getState().clearError();

    expect(useTopicStore.getState().error).toBeNull();
  });
});
