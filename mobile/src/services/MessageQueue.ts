// MessageQueue — offline-first message sending with retry logic
// Saves messages locally first (optimistic), then sends via API/WebSocket.
// Failed messages are retried with exponential backoff.

import * as messageRepo from '@/database/messageRepo';
import type { LocalMessage } from '@/database/types';

type PendingMessage = {
  localId: string;
  chatId: string;
  senderId: string;
  senderName: string;
  content: string;
  type: string;
  replyToId?: string;
  metadata?: Record<string, unknown>;
  retryCount: number;
};

type SendFunction = (msg: {
  chatId: string;
  content: string;
  type: string;
  replyToId?: string;
  metadata?: Record<string, unknown>;
}) => Promise<{ id: string }>;

const MAX_RETRIES = 5;
const MAX_BACKOFF_MS = 30000;

class MessageQueueService {
  private queue: PendingMessage[] = [];
  private isProcessing = false;
  private sendFn: SendFunction | null = null;
  private onStatusChange: ((localId: string, status: string, serverId?: string) => void) | null = null;

  /**
   * Set the send function (API or WebSocket).
   * Must be called before enqueue.
   */
  setSendFunction(fn: SendFunction): void {
    this.sendFn = fn;
  }

  /**
   * Set a callback for status changes.
   * Used to update UI (Zustand store, etc.).
   */
  setStatusChangeHandler(
    handler: (localId: string, status: string, serverId?: string) => void
  ): void {
    this.onStatusChange = handler;
  }

  /**
   * Enqueue a message for sending.
   * Saves to local DB immediately (optimistic insert).
   * Returns the local ID.
   */
  async enqueue(msg: Omit<PendingMessage, 'localId' | 'retryCount'>): Promise<string> {
    const localMessage: Omit<LocalMessage, 'id'> = {
      server_id: null,
      chat_id: msg.chatId,
      sender_id: msg.senderId,
      sender_name: msg.senderName,
      content: msg.content,
      type: msg.type,
      status: 'sending',
      reply_to_id: msg.replyToId ?? null,
      metadata: msg.metadata ? JSON.stringify(msg.metadata) : null,
      is_deleted: 0,
      is_pending: 1,
      created_at: Date.now(),
    };

    const localId = await messageRepo.insertMessage(localMessage);

    this.queue.push({
      localId,
      chatId: msg.chatId,
      senderId: msg.senderId,
      senderName: msg.senderName,
      content: msg.content,
      type: msg.type,
      replyToId: msg.replyToId,
      metadata: msg.metadata,
      retryCount: 0,
    });

    this.onStatusChange?.(localId, 'sending');
    this.processQueue();

    return localId;
  }

  /**
   * Process the queue — send messages one by one.
   */
  private async processQueue(): Promise<void> {
    if (this.isProcessing || this.queue.length === 0) return;
    this.isProcessing = true;

    while (this.queue.length > 0) {
      const msg = this.queue[0];
      if (!msg) break;

      try {
        if (!this.sendFn) {
          throw new Error('Send function not set');
        }

        const result = await this.sendFn({
          chatId: msg.chatId,
          content: msg.content,
          type: msg.type,
          replyToId: msg.replyToId,
          metadata: msg.metadata,
        });

        // Mark as sent in local DB
        await messageRepo.markMessageSent(msg.localId, result.id);
        this.onStatusChange?.(msg.localId, 'sent', result.id);
        this.queue.shift();
      } catch {
        if (msg.retryCount < MAX_RETRIES) {
          const backoffMs = this.getBackoff(msg.retryCount);
          msg.retryCount++;
          await this.delay(backoffMs);
        } else {
          // Max retries exceeded — mark as failed
          await messageRepo.markMessageFailed(msg.localId);
          this.onStatusChange?.(msg.localId, 'failed');
          this.queue.shift();
        }
      }
    }

    this.isProcessing = false;
  }

  /**
   * Retry a failed message.
   */
  async retry(localId: string): Promise<void> {
    // Reload from local DB
    const pending = await messageRepo.getPendingMessages();
    const msg = pending.find((m) => m.id === localId);

    if (!msg) return;

    // Re-enqueue
    this.queue.push({
      localId: msg.id,
      chatId: msg.chat_id,
      senderId: msg.sender_id,
      senderName: msg.sender_name,
      content: msg.content ?? '',
      type: msg.type,
      replyToId: msg.reply_to_id ?? undefined,
      metadata: msg.metadata ? JSON.parse(msg.metadata) : undefined,
      retryCount: 0,
    });

    this.onStatusChange?.(localId, 'sending');
    this.processQueue();
  }

  /**
   * Flush pending messages from local DB.
   * Called when coming back online.
   */
  async flushPending(): Promise<void> {
    const pending = await messageRepo.getPendingMessages();

    for (const msg of pending) {
      // Only add if not already in queue
      const alreadyQueued = this.queue.some((q) => q.localId === msg.id);
      if (!alreadyQueued) {
        this.queue.push({
          localId: msg.id,
          chatId: msg.chat_id,
          senderId: msg.sender_id,
          senderName: msg.sender_name,
          content: msg.content ?? '',
          type: msg.type,
          replyToId: msg.reply_to_id ?? undefined,
          metadata: msg.metadata ? JSON.parse(msg.metadata) : undefined,
          retryCount: 0,
        });
      }
    }

    if (this.queue.length > 0) {
      this.processQueue();
    }
  }

  /**
   * Get the number of pending messages in the queue.
   */
  getQueueLength(): number {
    return this.queue.length;
  }

  private getBackoff(retryCount: number): number {
    return Math.min(1000 * Math.pow(2, retryCount), MAX_BACKOFF_MS);
  }

  private delay(ms: number): Promise<void> {
    return new Promise((resolve) => setTimeout(resolve, ms));
  }
}

// Singleton instance
export const messageQueue = new MessageQueueService();
