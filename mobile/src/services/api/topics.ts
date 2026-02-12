// Topics API service (placeholder)
import apiClient from './client';

export const topicsApi = {
  create: (data: { parentType: string; parentId: string; name: string; memberIds: string[] }) =>
    apiClient.post('/topics', data),
  getById: (id: string) => apiClient.get(`/topics/${id}`),
  list: (parentType: string, parentId: string) =>
    apiClient.get(`/topics?parent_type=${parentType}&parent_id=${parentId}`),
};
