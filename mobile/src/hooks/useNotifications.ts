// useNotifications hook — initializes push notifications and handles deep linking
import { useEffect, useRef } from 'react';
import { useNavigation, CommonActions } from '@react-navigation/native';
import {
  pushNotificationService,
  type NotificationData,
} from '@/services/NotificationService';
import { useNotificationStore } from '@/stores/notificationStore';
import { useAuthStore } from '@/stores/authStore';

/**
 * Hook to initialize push notifications after authentication.
 * Should be called in a component that lives inside the navigation container.
 */
export function useNotifications() {
  const navigation = useNavigation();
  const isAuthenticated = useAuthStore((state) => state.isAuthenticated);
  const showNotification = useNotificationStore((state) => state.show);
  const initializedRef = useRef(false);

  useEffect(() => {
    if (!isAuthenticated) {
      // Cleanup on logout
      if (initializedRef.current) {
        pushNotificationService.cleanup();
        initializedRef.current = false;
      }
      return;
    }

    if (initializedRef.current) return;
    initializedRef.current = true;

    // Set deep link handler
    pushNotificationService.setNavigationHandler((data: NotificationData) => {
      handleDeepLink(navigation, data);
    });

    // Set foreground handler
    pushNotificationService.setForegroundHandler((notification) => {
      const { title, body } = notification.request.content;
      const data = notification.request.content.data as Record<string, string>;
      showNotification({
        title: title ?? '',
        body: body ?? '',
        data,
      });
    });

    // Initialize
    pushNotificationService.initialize();

    // Reset badge on app launch
    pushNotificationService.resetBadge();
  }, [isAuthenticated, navigation, showNotification]);
}

/** Route to the correct screen based on notification data */
function handleDeepLink(
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  navigation: any,
  data: NotificationData,
) {
  switch (data.type) {
    case 'message':
      if (data.chatId) {
        navigation.dispatch(
          CommonActions.navigate('ChatTab', {
            screen: 'Chat',
            params: { chatId: data.chatId, chatType: 'personal' },
          }),
        );
      }
      break;

    case 'group_message':
      if (data.chatId) {
        navigation.dispatch(
          CommonActions.navigate('ChatTab', {
            screen: 'Chat',
            params: { chatId: data.chatId, chatType: 'group' },
          }),
        );
      }
      break;

    case 'topic_message':
      if (data.topicId) {
        navigation.dispatch(
          CommonActions.navigate('ChatTab', {
            screen: 'Topic',
            params: { topicId: data.topicId },
          }),
        );
      }
      break;

    case 'signature_request':
    case 'document_locked':
      if (data.documentId) {
        navigation.dispatch(
          CommonActions.navigate('DocumentTab', {
            screen: 'DocumentEditor',
            params: { documentId: data.documentId },
          }),
        );
      }
      break;

    case 'group_invite':
      if (data.chatId) {
        navigation.dispatch(
          CommonActions.navigate('ChatTab', {
            screen: 'Chat',
            params: { chatId: data.chatId, chatType: 'group' },
          }),
        );
      }
      break;
  }
}
