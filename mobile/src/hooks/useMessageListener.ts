// useMessageListener — listens for real-time messages via WebSocket
import { useWSEvent } from './useWebSocket';
import { useMessageStore } from '@/stores/messageStore';
import { useChatStore } from '@/stores/chatStore';
import { useTopicStore } from '@/stores/topicStore';
import { wsClient } from '@/services/ws';
import type { Message, TopicMessage } from '@/types/chat';
import { useAuthStore } from '@/stores/authStore';
import { useCallback } from 'react';

type NewMessageEvent = {
  chatId: string;
  message: Message;
};

type MessageDeleteEvent = {
  type: string;
  payload: {
    chatId: string;
    messageId: string;
  };
};

type MessageStatusEvent = {
  chatId: string;
  messageId: string;
  userId: string;
  status: 'sent' | 'delivered' | 'read';
};

type NewTopicMessageEvent = {
  topicId: string;
  message: TopicMessage;
};

type TopicMessageDeleteEvent = {
  topicId: string;
  messageId: string;
};

/**
 * Listens for real-time message events and updates stores.
 * Call this once in the root app component.
 */
export function useMessageListener() {
  const addMessage = useMessageStore((s) => s.addMessage);
  const deleteMessage = useMessageStore((s) => s.deleteMessage);
  const updateLastMessage = useChatStore((s) => s.updateLastMessage);
  const currentUserId = useAuthStore((s) => s.user?.id);
  const addTopicMessage = useTopicStore((s) => s.addMessage);
  const deleteTopicMessage = useTopicStore((s) => s.deleteMessage);

  // New message received
  useWSEvent<NewMessageEvent>('new_message', useCallback((payload) => {
    if (!payload?.message) return;

    addMessage(payload.chatId, payload.message);
    updateLastMessage(payload.chatId, payload.message);

    // Send delivery ack if message is from someone else
    if (payload.message.senderId !== currentUserId) {
      wsClient.send('message_ack', {
        messageId: payload.message.id,
        chatId: payload.chatId,
        status: 'delivered',
      });
    }
  }, [addMessage, updateLastMessage, currentUserId]));

  // Message deleted
  useWSEvent<MessageDeleteEvent>('message_deleted', useCallback((payload) => {
    if (!payload?.payload) return;
    deleteMessage(payload.payload.chatId, payload.payload.messageId);
  }, [deleteMessage]));

  // Message status update (delivered / read)
  useWSEvent<MessageStatusEvent>('message_status', useCallback((_payload) => {
    // Status updates can be used to update message check marks
    // For now, we handle this at the UI level via the message model
    // A future enhancement would update the message status in the store
  }, []));

  // New topic message received
  useWSEvent<NewTopicMessageEvent>('new_topic_message', useCallback((payload) => {
    if (!payload?.message) return;
    addTopicMessage(payload.topicId, payload.message);
  }, [addTopicMessage]));

  // Topic message deleted
  useWSEvent<TopicMessageDeleteEvent>('topic_message_deleted', useCallback((payload) => {
    if (!payload?.topicId || !payload?.messageId) return;
    deleteTopicMessage(payload.topicId, payload.messageId);
  }, [deleteTopicMessage]));
}
