// Media API service (placeholder)
import apiClient from './client';

export const mediaApi = {
  upload: (formData: FormData) =>
    apiClient.post('/media/upload', formData, {
      headers: { 'Content-Type': 'multipart/form-data' },
    }),
  getById: (id: string) => apiClient.get(`/media/${id}`),
};
