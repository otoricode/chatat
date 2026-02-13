// useOnlineStatus — listens for real-time online/offline status via WebSocket
import { useCallback } from 'react';
import { useWSEvent } from './useWebSocket';
import { useContactStore } from '@/stores/contactStore';
import { useChatStore } from '@/stores/chatStore';

type OnlineStatusEvent = {
  userId: string;
  isOnline: boolean;
  lastSeen: string;
};

/**
 * Listens for online_status events and updates contact + chat stores.
 * Call this once in the root app component.
 */
export function useOnlineStatus() {
  const updateContactStatus = useContactStore((s) => s.updateOnlineStatus);
  const updateChatStatus = useChatStore((s) => s.updateChatOnlineStatus);

  useWSEvent<OnlineStatusEvent>('online_status', useCallback((payload) => {
    if (!payload?.userId) return;
    updateContactStatus(payload.userId, payload.isOnline, payload.lastSeen);
    updateChatStatus(payload.userId, payload.isOnline, payload.lastSeen);
  }, [updateContactStatus, updateChatStatus]));
}
