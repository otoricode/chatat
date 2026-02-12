// Entities API service (placeholder)
import apiClient from './client';

export const entitiesApi = {
  create: (data: { name: string; type: string }) =>
    apiClient.post('/entities', data),
  list: () => apiClient.get('/entities'),
  getById: (id: string) => apiClient.get(`/entities/${id}`),
};
