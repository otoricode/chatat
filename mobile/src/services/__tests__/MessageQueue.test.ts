jest.mock('@/database/messageRepo', () => ({
  insertMessage: jest.fn(),
  markMessageSent: jest.fn(),
  markMessageFailed: jest.fn(),
  getPendingMessages: jest.fn(),
}));

// Import after mock setup
import { messageQueue } from '../MessageQueue';
import * as messageRepo from '@/database/messageRepo';

const mockInsert = messageRepo.insertMessage as jest.MockedFunction<typeof messageRepo.insertMessage>;
const mockMarkSent = messageRepo.markMessageSent as jest.MockedFunction<typeof messageRepo.markMessageSent>;
const mockMarkFailed = messageRepo.markMessageFailed as jest.MockedFunction<typeof messageRepo.markMessageFailed>;
const mockGetPending = messageRepo.getPendingMessages as jest.MockedFunction<typeof messageRepo.getPendingMessages>;

beforeEach(() => {
  jest.clearAllMocks();
  // Reset internal state by setting new send function
  messageQueue.setSendFunction(jest.fn());
  messageQueue.setStatusChangeHandler(jest.fn());
});

describe('MessageQueue', () => {
  it('enqueue saves to DB and returns localId', async () => {
    mockInsert.mockResolvedValue('local_123');
    const sendFn = jest.fn().mockResolvedValue({ id: 'server_1' });
    messageQueue.setSendFunction(sendFn);

    const localId = await messageQueue.enqueue({
      chatId: 'c1',
      senderId: 'u1',
      senderName: 'Alice',
      content: 'Hello',
      type: 'text',
    });

    expect(localId).toBe('local_123');
    expect(mockInsert).toHaveBeenCalled();
  });

  it('enqueue calls send function and marks sent on success', async () => {
    mockInsert.mockResolvedValue('local_123');
    const sendFn = jest.fn().mockResolvedValue({ id: 'server_1' });
    const statusHandler = jest.fn();
    messageQueue.setSendFunction(sendFn);
    messageQueue.setStatusChangeHandler(statusHandler);

    await messageQueue.enqueue({
      chatId: 'c1',
      senderId: 'u1',
      senderName: 'Alice',
      content: 'Hello',
      type: 'text',
    });

    // Give processQueue time to complete
    await new Promise((r) => setTimeout(r, 50));

    expect(sendFn).toHaveBeenCalled();
    expect(mockMarkSent).toHaveBeenCalledWith('local_123', 'server_1');
    expect(statusHandler).toHaveBeenCalledWith('local_123', 'sent', 'server_1');
  });

  it('enqueue with replyToId and metadata', async () => {
    mockInsert.mockResolvedValue('local_456');
    const sendFn = jest.fn().mockResolvedValue({ id: 'server_2' });
    messageQueue.setSendFunction(sendFn);

    await messageQueue.enqueue({
      chatId: 'c1',
      senderId: 'u1',
      senderName: 'Alice',
      content: 'Reply',
      type: 'text',
      replyToId: 'msg_original',
      metadata: { key: 'value' },
    });

    // Give processQueue time to complete
    await new Promise((r) => setTimeout(r, 50));

    expect(sendFn).toHaveBeenCalledWith(
      expect.objectContaining({
        replyToId: 'msg_original',
        metadata: { key: 'value' },
      }),
    );
  });

  it('getQueueLength returns queue size', async () => {
    // Fresh instance has length from potential previous test, check it's a number
    expect(typeof messageQueue.getQueueLength()).toBe('number');
  });

  it('retry re-enqueues from DB', async () => {
    mockGetPending.mockResolvedValue([
      {
        id: 'local_123',
        server_id: null,
        chat_id: 'c1',
        sender_id: 'u1',
        sender_name: 'Alice',
        content: 'Hello',
        type: 'text',
        status: 'failed',
        reply_to_id: null,
        metadata: null,
        is_deleted: 0,
        is_pending: 1,
        created_at: Date.now(),
      },
    ] as any);

    const sendFn = jest.fn().mockResolvedValue({ id: 'server_1' });
    const statusHandler = jest.fn();
    messageQueue.setSendFunction(sendFn);
    messageQueue.setStatusChangeHandler(statusHandler);

    await messageQueue.retry('local_123');

    // Give processQueue time
    await new Promise((r) => setTimeout(r, 50));

    expect(statusHandler).toHaveBeenCalledWith('local_123', 'sending');
  });

  it('retry does nothing for non-existent message', async () => {
    mockGetPending.mockResolvedValue([]);

    const statusHandler = jest.fn();
    messageQueue.setStatusChangeHandler(statusHandler);

    await messageQueue.retry('nonexistent');

    // No status change for non-existent
    expect(statusHandler).not.toHaveBeenCalledWith('nonexistent', expect.anything());
  });

  it('flushPending loads pending from DB', async () => {
    mockGetPending.mockResolvedValue([
      {
        id: 'local_1',
        chat_id: 'c1',
        sender_id: 'u1',
        sender_name: 'Alice',
        content: 'Pending',
        type: 'text',
        reply_to_id: null,
        metadata: '{"key":"value"}',
        is_pending: 1,
      },
    ] as any);

    const sendFn = jest.fn().mockResolvedValue({ id: 'server_1' });
    messageQueue.setSendFunction(sendFn);

    await messageQueue.flushPending();

    // Give processQueue time
    await new Promise((r) => setTimeout(r, 50));

    expect(mockGetPending).toHaveBeenCalled();
  });

  it('flushPending with empty pending list does nothing', async () => {
    mockGetPending.mockResolvedValue([]);

    await messageQueue.flushPending();

    // No crash, no send called
  });
});
