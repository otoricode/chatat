// Chat store — manages chat list state
import { create } from 'zustand';
import { chatsApi } from '@/services/api/chats';
import type { ChatListItem, Message } from '@/types/chat';

type ChatState = {
  chats: ChatListItem[];
  isLoading: boolean;
  error: string | null;

  fetchChats: () => Promise<void>;
  updateLastMessage: (chatId: string, message: Message) => void;
  updateUnreadCount: (chatId: string, count: number) => void;
  pinChat: (chatId: string) => Promise<void>;
  unpinChat: (chatId: string) => Promise<void>;
  markAsRead: (chatId: string) => Promise<void>;
  clearError: () => void;
};

export const useChatStore = create<ChatState>()((set, get) => ({
  chats: [],
  isLoading: false,
  error: null,

  fetchChats: async () => {
    set({ isLoading: true, error: null });
    try {
      const res = await chatsApi.list();
      set({ chats: res.data.data ?? [], isLoading: false });
    } catch (err) {
      const msg = err instanceof Error ? err.message : 'Failed to load chats';
      set({ error: msg, isLoading: false });
    }
  },

  updateLastMessage: (chatId, message) => {
    set((state) => ({
      chats: state.chats.map((item) =>
        item.chat.id === chatId
          ? { ...item, lastMessage: message, unreadCount: item.unreadCount + 1 }
          : item,
      ),
    }));
  },

  updateUnreadCount: (chatId, count) => {
    set((state) => ({
      chats: state.chats.map((item) =>
        item.chat.id === chatId ? { ...item, unreadCount: count } : item,
      ),
    }));
  },

  pinChat: async (chatId) => {
    try {
      await chatsApi.pinChat(chatId);
      // Refetch to get updated order
      await get().fetchChats();
    } catch (err) {
      const msg = err instanceof Error ? err.message : 'Failed to pin chat';
      set({ error: msg });
    }
  },

  unpinChat: async (chatId) => {
    try {
      await chatsApi.unpinChat(chatId);
      await get().fetchChats();
    } catch (err) {
      const msg = err instanceof Error ? err.message : 'Failed to unpin chat';
      set({ error: msg });
    }
  },

  markAsRead: async (chatId) => {
    try {
      await chatsApi.markAsRead(chatId);
      set((state) => ({
        chats: state.chats.map((item) =>
          item.chat.id === chatId ? { ...item, unreadCount: 0 } : item,
        ),
      }));
    } catch {
      // Silent fail for mark as read
    }
  },

  clearError: () => set({ error: null }),
}));
