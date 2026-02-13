// Message store — manages messages per chat
import { create } from 'zustand';
import { chatsApi } from '@/services/api/chats';
import type { Message } from '@/types/chat';

type MessageState = {
  messages: Record<string, Message[]>;
  isLoading: boolean;
  hasMore: Record<string, boolean>;
  cursor: Record<string, string>;
  error: string | null;

  fetchMessages: (chatId: string) => Promise<void>;
  fetchMore: (chatId: string) => Promise<void>;
  addMessage: (chatId: string, message: Message) => void;
  deleteMessage: (chatId: string, messageId: string) => void;
  clearMessages: (chatId: string) => void;
  clearError: () => void;
};

export const useMessageStore = create<MessageState>()((set, get) => ({
  messages: {},
  isLoading: false,
  hasMore: {},
  cursor: {},
  error: null,

  fetchMessages: async (chatId) => {
    set({ isLoading: true, error: null });
    try {
      const res = await chatsApi.getMessages(chatId, undefined, 50);
      const data = res.data.data ?? [];
      const meta = res.data.meta;
      set((state) => ({
        messages: { ...state.messages, [chatId]: data },
        hasMore: { ...state.hasMore, [chatId]: meta.hasMore },
        cursor: { ...state.cursor, [chatId]: meta.cursor },
        isLoading: false,
      }));
    } catch (err) {
      const msg = err instanceof Error ? err.message : 'Failed to load messages';
      set({ error: msg, isLoading: false });
    }
  },

  fetchMore: async (chatId) => {
    const state = get();
    if (!state.hasMore[chatId] || state.isLoading) return;

    const currentCursor = state.cursor[chatId];
    if (!currentCursor) return;

    set({ isLoading: true });
    try {
      const res = await chatsApi.getMessages(chatId, currentCursor, 50);
      const data = res.data.data ?? [];
      const meta = res.data.meta;
      set((s) => ({
        messages: {
          ...s.messages,
          [chatId]: [...(s.messages[chatId] ?? []), ...data],
        },
        hasMore: { ...s.hasMore, [chatId]: meta.hasMore },
        cursor: { ...s.cursor, [chatId]: meta.cursor },
        isLoading: false,
      }));
    } catch (err) {
      const msg = err instanceof Error ? err.message : 'Failed to load more messages';
      set({ error: msg, isLoading: false });
    }
  },

  addMessage: (chatId, message) => {
    set((state) => ({
      messages: {
        ...state.messages,
        [chatId]: [message, ...(state.messages[chatId] ?? [])],
      },
    }));
  },

  deleteMessage: (chatId, messageId) => {
    set((state) => ({
      messages: {
        ...state.messages,
        [chatId]: (state.messages[chatId] ?? []).map((m) =>
          m.id === messageId ? { ...m, isDeleted: true, deletedForAll: true, content: '' } : m,
        ),
      },
    }));
  },

  clearMessages: (chatId) => {
    set((state) => {
      const newMessages = { ...state.messages };
      delete newMessages[chatId];
      const newHasMore = { ...state.hasMore };
      delete newHasMore[chatId];
      const newCursor = { ...state.cursor };
      delete newCursor[chatId];
      return { messages: newMessages, hasMore: newHasMore, cursor: newCursor };
    });
  },

  clearError: () => set({ error: null }),
}));
