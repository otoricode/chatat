// Media API service
import apiClient from './client';
import type { ApiResponse, MediaResponse } from '@/types/chat';

export type UploadOptions = {
  contextType?: string;
  contextId?: string;
  onProgress?: (progress: number) => void;
};

export const mediaApi = {
  upload: (
    uri: string,
    filename: string,
    mimeType: string,
    options?: UploadOptions,
  ) => {
    const formData = new FormData();
    formData.append('file', {
      uri,
      name: filename,
      type: mimeType,
    } as unknown as Blob);

    if (options?.contextType) {
      formData.append('contextType', options.contextType);
    }
    if (options?.contextId) {
      formData.append('contextId', options.contextId);
    }

    return apiClient.post<ApiResponse<MediaResponse>>('/media/upload', formData, {
      headers: { 'Content-Type': 'multipart/form-data' },
      onUploadProgress: (event) => {
        if (event.total && options?.onProgress) {
          options.onProgress(Math.round((event.loaded * 100) / event.total));
        }
      },
    });
  },

  getById: (id: string) =>
    apiClient.get<ApiResponse<MediaResponse>>(`/media/${id}`),

  getDownloadURL: (id: string) =>
    apiClient.get<string>(`/media/${id}/download`, {
      maxRedirects: 0,
      validateStatus: (status: number) => status >= 200 && status < 400,
    }),

  delete: (id: string) => apiClient.delete(`/media/${id}`),
};
