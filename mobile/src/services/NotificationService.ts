// Push notification service — handles token registration, permissions, and deep linking
import { Platform, AppState, type AppStateStatus } from 'react-native';
import * as Notifications from 'expo-notifications';
import { notificationsApi } from '@/services/api';

// Notification data shape from backend
export type NotificationData = {
  type: string;
  chatId?: string;
  topicId?: string;
  documentId?: string;
};

// Configure how notifications appear when app is in foreground
Notifications.setNotificationHandler({
  handleNotification: async () => ({
    shouldShowAlert: false, // We handle foreground with InAppNotification
    shouldPlaySound: true,
    shouldSetBadge: true,
    shouldShowBanner: false,
    shouldShowList: true,
  }),
});

class PushNotificationService {
  private token: string | null = null;
  private notificationListener: Notifications.EventSubscription | null = null;
  private responseListener: Notifications.EventSubscription | null = null;
  private tokenRefreshListener: Notifications.EventSubscription | null = null;
  private navigationHandler: ((data: NotificationData) => void) | null = null;
  private foregroundHandler: ((notification: Notifications.Notification) => void) | null = null;

  /** Initialize push notifications — call after login */
  async initialize(): Promise<void> {
    const { status: existingStatus } = await Notifications.getPermissionsAsync();
    let finalStatus = existingStatus;

    if (existingStatus !== 'granted') {
      const { status } = await Notifications.requestPermissionsAsync();
      finalStatus = status;
    }

    if (finalStatus !== 'granted') {
      console.warn('Push notification permission not granted');
      return;
    }

    // Get push token
    try {
      const tokenData = await Notifications.getExpoPushTokenAsync();
      this.token = tokenData.data;
      await this.registerToken(this.token);
    } catch (error) {
      console.warn('Failed to get push token:', error);
      return;
    }

    // Listen for token refresh
    this.tokenRefreshListener = Notifications.addPushTokenListener(async (newToken) => {
      const tokenString = newToken.data;
      if (tokenString !== this.token) {
        this.token = tokenString;
        await this.registerToken(tokenString);
      }
    });

    // Listen for incoming notifications (foreground)
    this.notificationListener = Notifications.addNotificationReceivedListener(
      (notification) => {
        if (this.foregroundHandler) {
          this.foregroundHandler(notification);
        }
      },
    );

    // Listen for notification taps (opens app)
    this.responseListener = Notifications.addNotificationResponseReceivedListener(
      (response) => {
        const data = response.notification.request.content.data as NotificationData;
        if (data && this.navigationHandler) {
          this.navigationHandler(data);
        }
      },
    );

    // Check if app was opened from a notification
    const lastResponse = await Notifications.getLastNotificationResponseAsync();
    if (lastResponse) {
      const data = lastResponse.notification.request.content.data as NotificationData;
      if (data && this.navigationHandler) {
        this.navigationHandler(data);
      }
    }
  }

  /** Set handler for deep linking from notification taps */
  setNavigationHandler(handler: (data: NotificationData) => void): void {
    this.navigationHandler = handler;
  }

  /** Set handler for foreground notifications */
  setForegroundHandler(handler: (notification: Notifications.Notification) => void): void {
    this.foregroundHandler = handler;
  }

  /** Cleanup — call on logout */
  async cleanup(): Promise<void> {
    if (this.token) {
      try {
        await notificationsApi.unregisterDevice(this.token);
      } catch {
        // Ignore unregister errors on cleanup
      }
    }

    this.notificationListener?.remove();
    this.responseListener?.remove();
    this.tokenRefreshListener?.remove();
    this.notificationListener = null;
    this.responseListener = null;
    this.tokenRefreshListener = null;
    this.token = null;
    this.navigationHandler = null;
    this.foregroundHandler = null;
  }

  /** Register device token with backend */
  private async registerToken(token: string): Promise<void> {
    try {
      const platform = Platform.OS as 'ios' | 'android';
      await notificationsApi.registerDevice(token, platform);
    } catch (error) {
      console.warn('Failed to register device token:', error);
    }
  }

  /** Get the current push token */
  getToken(): string | null {
    return this.token;
  }

  /** Reset badge count */
  async resetBadge(): Promise<void> {
    await Notifications.setBadgeCountAsync(0);
  }
}

export const pushNotificationService = new PushNotificationService();
