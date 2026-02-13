// useTypingIndicator — manages typing indicators via WebSocket
import { useState, useEffect, useCallback, useRef } from 'react';
import { useWSEvent, useWSSend } from './useWebSocket';

type TypingEvent = {
  chatId: string;
  userId: string;
  userName: string;
  isTyping: boolean;
};

type TypingUser = {
  userId: string;
  userName: string;
  expiresAt: number;
};

/**
 * Returns typing users for a given chat, and a function to send typing events.
 */
export function useTypingIndicator(chatId: string) {
  const [typingUsers, setTypingUsers] = useState<TypingUser[]>([]);
  const send = useWSSend();
  const lastSentRef = useRef(0);
  const stopTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null);

  // Listen for typing events
  useWSEvent<TypingEvent>('typing', useCallback((payload) => {
    if (payload.chatId !== chatId) return;

    setTypingUsers((prev) => {
      if (payload.isTyping) {
        // Add or update typing user
        const existing = prev.filter((u) => u.userId !== payload.userId);
        return [
          ...existing,
          {
            userId: payload.userId,
            userName: payload.userName,
            expiresAt: Date.now() + 3000,
          },
        ];
      } else {
        // Remove typing user
        return prev.filter((u) => u.userId !== payload.userId);
      }
    });
  }, [chatId]));

  // Auto-clear expired typing indicators
  useEffect(() => {
    const interval = setInterval(() => {
      const now = Date.now();
      setTypingUsers((prev) => {
        const filtered = prev.filter((u) => u.expiresAt > now);
        if (filtered.length !== prev.length) return filtered;
        return prev;
      });
    }, 1000);

    return () => clearInterval(interval);
  }, []);

  // Send typing event (debounced: max 1 per 2s)
  const sendTyping = useCallback(
    (isTyping: boolean) => {
      const now = Date.now();
      if (isTyping && now - lastSentRef.current < 2000) return;

      lastSentRef.current = now;
      send('typing', { chatId, isTyping });

      // Auto-stop after 2s of no calls
      if (stopTimerRef.current) {
        clearTimeout(stopTimerRef.current);
      }
      if (isTyping) {
        stopTimerRef.current = setTimeout(() => {
          send('typing', { chatId, isTyping: false });
        }, 2000);
      }
    },
    [chatId, send],
  );

  // Cleanup on unmount
  useEffect(() => {
    return () => {
      if (stopTimerRef.current) {
        clearTimeout(stopTimerRef.current);
      }
    };
  }, []);

  // Format typing text
  const typingText = formatTypingText(typingUsers);

  return { typingUsers, typingText, sendTyping };
}

function formatTypingText(users: TypingUser[]): string | null {
  if (users.length === 0) return null;
  if (users.length === 1) {
    const name = users[0]?.userName || 'Seseorang';
    return `${name} sedang mengetik...`;
  }
  if (users.length === 2) {
    const n1 = users[0]?.userName || 'Seseorang';
    const n2 = users[1]?.userName || 'Seseorang';
    return `${n1}, ${n2} sedang mengetik...`;
  }
  const n1 = users[0]?.userName || 'Seseorang';
  return `${n1} dan ${users.length - 1} lainnya sedang mengetik...`;
}
