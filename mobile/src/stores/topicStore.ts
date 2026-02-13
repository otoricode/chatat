// Topic store — manages topic list and messages state
import { create } from 'zustand';
import { topicsApi } from '@/services/api/topics';
import type { TopicListItem, TopicMessage } from '@/types/chat';

type TopicState = {
  // topics by chatId
  topicsByChat: Record<string, TopicListItem[]>;
  // messages by topicId
  messagesByTopic: Record<string, TopicMessage[]>;
  isLoading: boolean;
  error: string | null;

  fetchTopics: (chatId: string) => Promise<void>;
  addTopic: (chatId: string, topic: TopicListItem) => void;
  removeTopic: (chatId: string, topicId: string) => void;

  fetchMessages: (topicId: string) => Promise<void>;
  addMessage: (topicId: string, message: TopicMessage) => void;
  deleteMessage: (topicId: string, messageId: string) => void;
  clearError: () => void;
};

export const useTopicStore = create<TopicState>()((set, get) => ({
  topicsByChat: {},
  messagesByTopic: {},
  isLoading: false,
  error: null,

  fetchTopics: async (chatId) => {
    set({ isLoading: true, error: null });
    try {
      const res = await topicsApi.listByChat(chatId);
      set((state) => ({
        topicsByChat: {
          ...state.topicsByChat,
          [chatId]: res.data.data ?? [],
        },
        isLoading: false,
      }));
    } catch (err) {
      const msg = err instanceof Error ? err.message : 'Failed to load topics';
      set({ error: msg, isLoading: false });
    }
  },

  addTopic: (chatId, topic) => {
    set((state) => ({
      topicsByChat: {
        ...state.topicsByChat,
        [chatId]: [topic, ...(state.topicsByChat[chatId] ?? [])],
      },
    }));
  },

  removeTopic: (chatId, topicId) => {
    set((state) => ({
      topicsByChat: {
        ...state.topicsByChat,
        [chatId]: (state.topicsByChat[chatId] ?? []).filter(
          (t) => t.topic.id !== topicId,
        ),
      },
    }));
  },

  fetchMessages: async (topicId) => {
    set({ isLoading: true, error: null });
    try {
      const res = await topicsApi.getMessages(topicId);
      set((state) => ({
        messagesByTopic: {
          ...state.messagesByTopic,
          [topicId]: res.data.data ?? [],
        },
        isLoading: false,
      }));
    } catch (err) {
      const msg = err instanceof Error ? err.message : 'Failed to load messages';
      set({ error: msg, isLoading: false });
    }
  },

  addMessage: (topicId, message) => {
    set((state) => {
      const existing = state.messagesByTopic[topicId] ?? [];
      // Avoid duplicates
      if (existing.some((m) => m.id === message.id)) return state;
      return {
        messagesByTopic: {
          ...state.messagesByTopic,
          [topicId]: [message, ...existing],
        },
      };
    });
  },

  deleteMessage: (topicId, messageId) => {
    set((state) => ({
      messagesByTopic: {
        ...state.messagesByTopic,
        [topicId]: (state.messagesByTopic[topicId] ?? []).map((m) =>
          m.id === messageId ? { ...m, isDeleted: true, content: '' } : m,
        ),
      },
    }));
  },

  clearError: () => set({ error: null }),
}));
