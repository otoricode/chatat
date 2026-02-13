// useWebSocket — manages WebSocket connection lifecycle
import { useEffect, useCallback, useRef } from 'react';
import { wsClient } from '@/services/ws';
import { useAuthStore } from '@/stores/authStore';
import { WS_BASE_URL } from '@/lib/constants';
import type { ConnectionState } from '@/services/ws';

/**
 * Connects the WebSocket client when the user has an access token.
 * Disconnects on logout or unmount.
 * Call this once in the root app component.
 */
export function useWebSocket() {
  const accessToken = useAuthStore((s) => s.accessToken);

  useEffect(() => {
    if (accessToken) {
      wsClient.connect(WS_BASE_URL, accessToken);
    }
    return () => {
      wsClient.disconnect();
    };
  }, [accessToken]);
}

/**
 * Subscribe to a specific WS event type.
 * Handler is automatically cleaned up on unmount.
 */
export function useWSEvent<T = unknown>(type: string, handler: (payload: T) => void) {
  const handlerRef = useRef(handler);
  handlerRef.current = handler;

  useEffect(() => {
    const stableHandler = (payload: unknown) => {
      handlerRef.current(payload as T);
    };
    return wsClient.on(type, stableHandler);
  }, [type]);
}

/**
 * Subscribe to WebSocket connection state changes.
 */
export function useWSConnectionState(handler: (state: ConnectionState) => void) {
  const handlerRef = useRef(handler);
  handlerRef.current = handler;

  useEffect(() => {
    const stableHandler = (state: ConnectionState) => {
      handlerRef.current(state);
    };
    return wsClient.onStateChange(stableHandler);
  }, []);
}

/**
 * Returns a function to send messages via WebSocket.
 */
export function useWSSend() {
  return useCallback((type: string, payload: unknown) => {
    wsClient.send(type, payload);
  }, []);
}
