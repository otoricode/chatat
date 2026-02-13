// Topics API service
import apiClient from './client';
import type {
  ApiResponse,
  PaginatedResponse,
  Topic,
  TopicListItem,
  TopicDetail,
  TopicMessage,
} from '@/types/chat';

export type CreateTopicRequest = {
  name: string;
  icon: string;
  description?: string;
  parentId: string;
  memberIds?: string[];
};

export type UpdateTopicRequest = {
  name?: string;
  icon?: string;
  description?: string;
};

export type SendTopicMessageRequest = {
  content: string;
  replyToId?: string | null;
  type?: string;
};

export const topicsApi = {
  list: () => apiClient.get<ApiResponse<TopicListItem[]>>('/topics'),

  listByChat: (chatId: string) =>
    apiClient.get<ApiResponse<TopicListItem[]>>(`/chats/${chatId}/topics`),

  create: (data: CreateTopicRequest) =>
    apiClient.post<ApiResponse<Topic>>('/topics', data),

  getById: (id: string) =>
    apiClient.get<ApiResponse<TopicDetail>>(`/topics/${id}`),

  update: (id: string, data: UpdateTopicRequest) =>
    apiClient.put<ApiResponse<Topic>>(`/topics/${id}`, data),

  delete: (id: string) => apiClient.delete(`/topics/${id}`),

  addMember: (topicId: string, userId: string) =>
    apiClient.post(`/topics/${topicId}/members`, { userId }),

  removeMember: (topicId: string, userId: string) =>
    apiClient.delete(`/topics/${topicId}/members/${userId}`),

  getMessages: (topicId: string, cursor?: string, limit?: number) => {
    const params = new URLSearchParams();
    if (cursor) params.set('cursor', cursor);
    if (limit) params.set('limit', String(limit));
    const qs = params.toString();
    return apiClient.get<PaginatedResponse<TopicMessage[]>>(
      `/topics/${topicId}/messages${qs ? `?${qs}` : ''}`,
    );
  },

  sendMessage: (topicId: string, data: SendTopicMessageRequest) =>
    apiClient.post<ApiResponse<TopicMessage>>(`/topics/${topicId}/messages`, data),

  deleteMessage: (topicId: string, messageId: string, forAll = false) =>
    apiClient.delete(`/topics/${topicId}/messages/${messageId}?forAll=${forAll}`),
};
