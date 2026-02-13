// Notification API endpoints
import apiClient from './client';

export const notificationsApi = {
  /** Register a device token for push notifications */
  registerDevice: (token: string, platform: 'ios' | 'android') =>
    apiClient.post('/notifications/devices', { token, platform }),

  /** Unregister a device token */
  unregisterDevice: (token: string) =>
    apiClient.delete('/notifications/devices', { data: { token } }),
};
