// Chats API service
import apiClient from './client';
import type {
  ApiResponse,
  PaginatedResponse,
  ChatListItem,
  ChatDetail,
  Chat,
  Message,
} from '@/types/chat';

export type SendMessageRequest = {
  content: string;
  replyToId?: string | null;
  type?: string;
  metadata?: Record<string, unknown>;
};

export type ForwardMessageRequest = {
  targetChatId: string;
};

export const chatsApi = {
  list: () => apiClient.get<ApiResponse<ChatListItem[]>>('/chats'),

  create: (contactId: string) =>
    apiClient.post<ApiResponse<Chat>>('/chats', { contactId }),

  getById: (id: string) => apiClient.get<ApiResponse<ChatDetail>>(`/chats/${id}`),

  pinChat: (id: string) => apiClient.put(`/chats/${id}/pin`),

  unpinChat: (id: string) => apiClient.delete(`/chats/${id}/pin`),

  markAsRead: (id: string) => apiClient.post(`/chats/${id}/read`),

  getMessages: (chatId: string, cursor?: string, limit?: number) => {
    const params = new URLSearchParams();
    if (cursor) params.set('cursor', cursor);
    if (limit) params.set('limit', String(limit));
    const qs = params.toString();
    return apiClient.get<PaginatedResponse<Message[]>>(
      `/chats/${chatId}/messages${qs ? `?${qs}` : ''}`,
    );
  },

  sendMessage: (chatId: string, data: SendMessageRequest) =>
    apiClient.post<ApiResponse<Message>>(`/chats/${chatId}/messages`, data),

  deleteMessage: (chatId: string, messageId: string, forAll = false) =>
    apiClient.delete(`/chats/${chatId}/messages/${messageId}?forAll=${forAll}`),

  forwardMessage: (chatId: string, messageId: string, data: ForwardMessageRequest) =>
    apiClient.post<ApiResponse<Message>>(
      `/chats/${chatId}/messages/${messageId}/forward`,
      data,
    ),

  searchMessages: (chatId: string, query: string) =>
    apiClient.get<ApiResponse<Message[]>>(
      `/chats/${chatId}/messages/search?q=${encodeURIComponent(query)}`,
    ),
};
