// User API service (placeholder)
import apiClient from './client';

type User = {
  id: string;
  name: string;
  phone: string;
  avatar: string;
  status: string;
};

type SetupProfileRequest = {
  name: string;
  avatar: string;
};

export const usersApi = {
  getMe: () => apiClient.get<User>('/users/me'),

  updateMe: (data: Partial<User>) =>
    apiClient.put<User>('/users/me', data),

  setupProfile: (data: SetupProfileRequest) =>
    apiClient.post<User>('/users/me/setup', data),

  deleteAccount: () => apiClient.delete('/users/me'),
};
