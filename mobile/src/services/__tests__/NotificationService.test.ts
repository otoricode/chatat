// Mock dependencies
jest.mock('expo-notifications', () => ({
  getPermissionsAsync: jest.fn(),
  requestPermissionsAsync: jest.fn(),
  getExpoPushTokenAsync: jest.fn(),
  getLastNotificationResponseAsync: jest.fn(),
  addNotificationReceivedListener: jest.fn(() => ({ remove: jest.fn() })),
  addNotificationResponseReceivedListener: jest.fn(() => ({ remove: jest.fn() })),
  addPushTokenListener: jest.fn(() => ({ remove: jest.fn() })),
  setNotificationHandler: jest.fn(),
  setBadgeCountAsync: jest.fn(),
}));

jest.mock('@/services/api', () => ({
  notificationsApi: {
    registerDevice: jest.fn().mockResolvedValue({}),
    unregisterDevice: jest.fn().mockResolvedValue({}),
  },
}));

jest.mock('react-native', () => ({
  Platform: { OS: 'ios' },
  AppState: { addEventListener: jest.fn() },
}));

import * as Notifications from 'expo-notifications';
import { notificationsApi } from '@/services/api';

// We need to import after mocks
const { pushNotificationService } = require('../NotificationService');

beforeEach(() => {
  jest.clearAllMocks();
});

describe('PushNotificationService', () => {
  describe('initialize', () => {
    it('requests permission and registers token when granted', async () => {
      (Notifications.getPermissionsAsync as jest.Mock).mockResolvedValue({ status: 'granted' });
      (Notifications.getExpoPushTokenAsync as jest.Mock).mockResolvedValue({ data: 'expo-token-123' });
      (Notifications.getLastNotificationResponseAsync as jest.Mock).mockResolvedValue(null);

      await pushNotificationService.initialize();

      expect(notificationsApi.registerDevice).toHaveBeenCalledWith('expo-token-123', 'ios');
    });

    it('requests permission if not granted', async () => {
      (Notifications.getPermissionsAsync as jest.Mock).mockResolvedValue({ status: 'undetermined' });
      (Notifications.requestPermissionsAsync as jest.Mock).mockResolvedValue({ status: 'granted' });
      (Notifications.getExpoPushTokenAsync as jest.Mock).mockResolvedValue({ data: 'expo-token' });
      (Notifications.getLastNotificationResponseAsync as jest.Mock).mockResolvedValue(null);

      await pushNotificationService.initialize();

      expect(Notifications.requestPermissionsAsync).toHaveBeenCalled();
    });

    it('returns early if permission denied', async () => {
      (Notifications.getPermissionsAsync as jest.Mock).mockResolvedValue({ status: 'denied' });
      (Notifications.requestPermissionsAsync as jest.Mock).mockResolvedValue({ status: 'denied' });

      await pushNotificationService.initialize();

      expect(Notifications.getExpoPushTokenAsync).not.toHaveBeenCalled();
    });

    it('handles push token error', async () => {
      (Notifications.getPermissionsAsync as jest.Mock).mockResolvedValue({ status: 'granted' });
      (Notifications.getExpoPushTokenAsync as jest.Mock).mockRejectedValue(new Error('token error'));

      await pushNotificationService.initialize();
      // Should not throw, just warns
    });
  });

  describe('setNavigationHandler', () => {
    it('stores handler', () => {
      const handler = jest.fn();
      pushNotificationService.setNavigationHandler(handler);
      expect((pushNotificationService as any).navigationHandler).toBe(handler);
    });
  });

  describe('setForegroundHandler', () => {
    it('stores handler', () => {
      const handler = jest.fn();
      pushNotificationService.setForegroundHandler(handler);
      expect((pushNotificationService as any).foregroundHandler).toBe(handler);
    });
  });

  describe('getToken', () => {
    it('returns token', () => {
      (pushNotificationService as any).token = 'my-token';
      expect(pushNotificationService.getToken()).toBe('my-token');
    });

    it('returns null if no token', () => {
      (pushNotificationService as any).token = null;
      expect(pushNotificationService.getToken()).toBeNull();
    });
  });

  describe('resetBadge', () => {
    it('calls setBadgeCountAsync', async () => {
      await pushNotificationService.resetBadge();
      expect(Notifications.setBadgeCountAsync).toHaveBeenCalledWith(0);
    });
  });

  describe('cleanup', () => {
    it('unregisters device and cleans up listeners', async () => {
      (pushNotificationService as any).token = 'tok';
      
      await pushNotificationService.cleanup();

      expect(notificationsApi.unregisterDevice).toHaveBeenCalledWith('tok');
      expect((pushNotificationService as any).token).toBeNull();
      expect((pushNotificationService as any).navigationHandler).toBeNull();
      expect((pushNotificationService as any).foregroundHandler).toBeNull();
    });

    it('handles unregister error gracefully', async () => {
      (pushNotificationService as any).token = 'tok';
      (notificationsApi.unregisterDevice as jest.Mock).mockRejectedValue(new Error('fail'));

      // Should not throw
      await pushNotificationService.cleanup();
    });
  });
});
