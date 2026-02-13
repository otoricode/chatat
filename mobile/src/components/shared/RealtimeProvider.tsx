// RealtimeProvider — initializes WebSocket connection and real-time listeners
// Wrap this around the main app content (inside auth check)
import React from 'react';
import { useWebSocket } from '@/hooks/useWebSocket';
import { useMessageListener } from '@/hooks/useMessageListener';
import { useOnlineStatus } from '@/hooks/useOnlineStatus';

type Props = {
  children: React.ReactNode;
};

export function RealtimeProvider({ children }: Props) {
  // Connect WebSocket when authenticated
  useWebSocket();

  // Listen for real-time messages
  useMessageListener();

  // Listen for online/offline status
  useOnlineStatus();

  return <>{children}</>;
}
