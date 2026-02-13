// Chats API service
import apiClient from './client';
import type {
  ApiResponse,
  PaginatedResponse,
  ChatListItem,
  ChatDetail,
  Chat,
  Message,
  GroupInfo,
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

export type CreateGroupRequest = {
  type: 'group';
  name: string;
  icon: string;
  description?: string;
  memberIds: string[];
};

export type UpdateGroupRequest = {
  name?: string;
  icon?: string;
  description?: string;
};

export const chatsApi = {
  list: () => apiClient.get<ApiResponse<ChatListItem[]>>('/chats'),

  create: (contactId: string) =>
    apiClient.post<ApiResponse<Chat>>('/chats', { contactId }),

  createGroup: (data: CreateGroupRequest) =>
    apiClient.post<ApiResponse<Chat>>('/chats', data),

  getById: (id: string) => apiClient.get<ApiResponse<ChatDetail>>(`/chats/${id}`),

  updateGroup: (id: string, data: UpdateGroupRequest) =>
    apiClient.put<ApiResponse<Chat>>(`/chats/${id}`, data),

  deleteGroup: (id: string) => apiClient.delete(`/chats/${id}`),

  getGroupInfo: (id: string) =>
    apiClient.get<ApiResponse<GroupInfo>>(`/chats/${id}/info`),

  pinChat: (id: string) => apiClient.put(`/chats/${id}/pin`),

  unpinChat: (id: string) => apiClient.delete(`/chats/${id}/pin`),

  markAsRead: (id: string) => apiClient.post(`/chats/${id}/read`),

  leaveGroup: (id: string) => apiClient.post(`/chats/${id}/leave`),

  addMember: (chatId: string, userId: string) =>
    apiClient.post(`/chats/${chatId}/members`, { userId }),

  removeMember: (chatId: string, userId: string) =>
    apiClient.delete(`/chats/${chatId}/members/${userId}`),

  promoteToAdmin: (chatId: string, userId: string) =>
    apiClient.put(`/chats/${chatId}/members/${userId}/admin`),

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
