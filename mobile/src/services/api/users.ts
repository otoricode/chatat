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

export type PrivacySettings = {
  lastSeenVisibility: 'everyone' | 'contacts' | 'nobody';
  onlineVisibility: 'everyone' | 'contacts' | 'nobody';
  readReceipts: boolean;
  profilePhotoVisibility: 'everyone' | 'contacts' | 'nobody';
};

export const usersApi = {
  getMe: () => apiClient.get<User>('/users/me'),

  updateMe: (data: Partial<User>) =>
    apiClient.put<User>('/users/me', data),

  setupProfile: (data: SetupProfileRequest) =>
    apiClient.post<User>('/users/me/setup', data),

  deleteAccount: () => apiClient.delete('/users/me'),

  getPrivacy: () =>
    apiClient.get<PrivacySettings>('/users/me/privacy'),

  updatePrivacy: (data: Partial<PrivacySettings>) =>
    apiClient.put<PrivacySettings>('/users/me/privacy', data),
};
