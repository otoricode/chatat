// Chats API service (placeholder)
import apiClient from './client';

export const chatsApi = {
  list: () => apiClient.get('/chats'),
  getById: (id: string) => apiClient.get(`/chats/${id}`),
  create: (data: { memberIds: string[] }) => apiClient.post('/chats', data),
  delete: (id: string) => apiClient.delete(`/chats/${id}`),
  getMessages: (chatId: string, cursor?: string) =>
    apiClient.get(`/chats/${chatId}/messages${cursor ? `?cursor=${cursor}` : ''}`),
  sendMessage: (chatId: string, data: { text: string; replyTo?: string }) =>
    apiClient.post(`/chats/${chatId}/messages`, data),
};
