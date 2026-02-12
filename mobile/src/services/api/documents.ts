// Documents API service (placeholder)
import apiClient from './client';

export const documentsApi = {
  create: (data: { title: string; contextType?: string; contextId?: string }) =>
    apiClient.post('/documents', data),
  getById: (id: string) => apiClient.get(`/documents/${id}`),
  update: (id: string, data: unknown) => apiClient.put(`/documents/${id}`, data),
  lock: (id: string) => apiClient.post(`/documents/${id}/lock`),
  sign: (id: string) => apiClient.post(`/documents/${id}/sign`),
};
