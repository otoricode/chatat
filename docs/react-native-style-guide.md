# React Native Style Guide

> **React Native Version:** 0.75+
> **Language:** TypeScript (strict mode)
> **Formatter:** Prettier
> **Linter:** ESLint
> **State Management:** Zustand
> **Navigation:** React Navigation
> **Styling:** StyleSheet.create (no styled-components)

---

## General Rules

### Formatting
- Gunakan Prettier dengan konfigurasi project
- Max line width: 100 characters
- Indentation: 2 spaces
- Single quotes, trailing commas
- Semicolons: yes

### ESLint
- Extend `@react-native/eslint-config`
- No `any` type — zero tolerance
- No unused variables/imports
- Prefer `const` over `let`

### TypeScript
- Strict mode enabled
- No `any` — gunakan `unknown` jika benar-benar tidak tahu tipe
- Explicit return types pada exported functions
- Prefer `type` over `interface` kecuali perlu extends

---

## Component Structure

### Functional Components Only

```typescript
// components/chat/MessageBubble.tsx

import React from 'react';
import { View, Text, StyleSheet } from 'react-native';
import type { Message } from '@/types/message';

type MessageBubbleProps = {
  message: Message;
  isMine: boolean;
  onLongPress?: (message: Message) => void;
};

export function MessageBubble({ message, isMine, onLongPress }: MessageBubbleProps) {
  return (
    <View style={[styles.bubble, isMine ? styles.mine : styles.theirs]}>
      <Text style={styles.text}>{message.text}</Text>
      <Text style={styles.time}>{formatTime(message.createdAt)}</Text>
    </View>
  );
}

const styles = StyleSheet.create({
  bubble: {
    maxWidth: '75%',
    paddingHorizontal: 12,
    paddingVertical: 8,
    borderRadius: 16,
    marginVertical: 2,
  },
  mine: {
    alignSelf: 'flex-end',
    backgroundColor: '#005C4B',
  },
  theirs: {
    alignSelf: 'flex-start',
    backgroundColor: '#1F2C34',
  },
  text: {
    color: '#E9EDEF',
    fontSize: 16,
  },
  time: {
    color: '#8696A0',
    fontSize: 11,
    alignSelf: 'flex-end',
    marginTop: 4,
  },
});
```

### Component File Order

```typescript
// 1. Imports (React, RN, external, internal, types)
// 2. Type definitions (Props)
// 3. Component function (exported)
// 4. Sub-components (unexported, if small)
// 5. StyleSheet.create
// 6. Helper functions (if co-located)
```

---

## Screen Pattern

```typescript
// screens/chat/ChatScreen.tsx

import React, { useCallback } from 'react';
import { SafeAreaView, StyleSheet } from 'react-native';
import type { NativeStackScreenProps } from '@react-navigation/native-stack';
import type { MainStackParamList } from '@/navigation/types';
import { useChat } from '@/hooks/useChat';
import { useAuth } from '@/hooks/useAuth';
import { ChatView } from '@/components/chat/ChatView';
import { LoadingState } from '@/components/shared/LoadingState';
import { ErrorState } from '@/components/shared/ErrorState';

type Props = NativeStackScreenProps<MainStackParamList, 'Chat'>;

export function ChatScreen({ route }: Props) {
  const { chatId } = route.params;
  const { messages, isLoading, error, sendMessage, refresh } = useChat(chatId);
  const { user } = useAuth();

  const handleSend = useCallback((text: string) => {
    sendMessage(text);
  }, [sendMessage]);

  if (isLoading) return <LoadingState />;
  if (error) return <ErrorState error={error} onRetry={refresh} />;

  return (
    <SafeAreaView style={styles.container}>
      <ChatView
        messages={messages}
        currentUserId={user.id}
        onSend={handleSend}
      />
    </SafeAreaView>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: '#0B141A',
  },
});
```

---

## Hooks

### Custom Hook Pattern

```typescript
// hooks/useChat.ts

import { useState, useCallback, useEffect } from 'react';
import { chatApi } from '@/services/chatApi';
import { parseError } from '@/types/api';
import type { Message } from '@/types/message';
import type { ApiError } from '@/types/api';

export function useChat(chatId: string) {
  const [messages, setMessages] = useState<Message[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<ApiError | null>(null);

  const fetchMessages = useCallback(async () => {
    setIsLoading(true);
    setError(null);
    try {
      const data = await chatApi.getMessages(chatId);
      setMessages(data);
    } catch (err) {
      setError(parseError(err));
    } finally {
      setIsLoading(false);
    }
  }, [chatId]);

  useEffect(() => {
    fetchMessages();
  }, [fetchMessages]);

  const sendMessage = useCallback(async (text: string) => {
    try {
      const msg = await chatApi.sendMessage(chatId, { text });
      setMessages((prev) => [msg, ...prev]);
    } catch (err) {
      setError(parseError(err));
    }
  }, [chatId]);

  return { messages, isLoading, error, sendMessage, refresh: fetchMessages };
}
```

### Hook Rules
- Prefix with `use`
- Return object with named properties (not array)
- Handle loading, error, data states internally
- Memoize callbacks with `useCallback`
- Clean up effects properly

---

## State Management (Zustand)

```typescript
// stores/authStore.ts

import { create } from 'zustand';
import { persist, createJSONStorage } from 'zustand/middleware';
import AsyncStorage from '@react-native-async-storage/async-storage';
import type { User } from '@/types/user';

type AuthState = {
  user: User | null;
  token: string | null;
  isAuthenticated: boolean;

  setAuth: (user: User, token: string) => void;
  logout: () => void;
  updateProfile: (updates: Partial<User>) => void;
};

export const useAuthStore = create<AuthState>()(
  persist(
    (set) => ({
      user: null,
      token: null,
      isAuthenticated: false,

      setAuth: (user, token) =>
        set({ user, token, isAuthenticated: true }),

      logout: () =>
        set({ user: null, token: null, isAuthenticated: false }),

      updateProfile: (updates) =>
        set((state) => ({
          user: state.user ? { ...state.user, ...updates } : null,
        })),
    }),
    {
      name: 'auth-storage',
      storage: createJSONStorage(() => AsyncStorage),
    }
  )
);
```

### Store Rules
- One store per domain
- Keep stores flat (no deep nesting)
- Use `persist` middleware for data that survives app restart
- Actions are defined inside the store (not external)
- Use selectors for derived data

---

## Navigation

```typescript
// navigation/types.ts

export type AuthStackParamList = {
  Login: undefined;
  OTP: { phone: string };
  ProfileSetup: undefined;
};

export type MainStackParamList = {
  ChatList: undefined;
  Chat: { chatId: string };
  ChatInfo: { chatId: string };
  NewChat: undefined;
  NewGroup: undefined;
  Topic: { topicId: string };
  Document: { documentId: string };
  Settings: undefined;
};

// navigation/RootNavigator.tsx

export function RootNavigator() {
  const { isAuthenticated } = useAuthStore();

  return (
    <NavigationContainer>
      {isAuthenticated ? <MainNavigator /> : <AuthNavigator />}
    </NavigationContainer>
  );
}
```

---

## API Services

```typescript
// services/api.ts

import { useAuthStore } from '@/stores/authStore';

const BASE_URL = 'https://api.chatat.app/v1';

async function request<T>(
  path: string,
  options: RequestInit = {}
): Promise<T> {
  const token = useAuthStore.getState().token;

  const response = await fetch(`${BASE_URL}${path}`, {
    ...options,
    headers: {
      'Content-Type': 'application/json',
      ...(token ? { Authorization: `Bearer ${token}` } : {}),
      ...options.headers,
    },
  });

  if (!response.ok) {
    const error = await response.json();
    throw error;
  }

  return response.json();
}

export const api = {
  get: <T>(path: string) => request<T>(path),
  post: <T>(path: string, body: unknown) =>
    request<T>(path, { method: 'POST', body: JSON.stringify(body) }),
  put: <T>(path: string, body: unknown) =>
    request<T>(path, { method: 'PUT', body: JSON.stringify(body) }),
  delete: <T>(path: string) =>
    request<T>(path, { method: 'DELETE' }),
};

// services/chatApi.ts

import { api } from './api';
import type { Chat, ChatSummary } from '@/types/chat';
import type { Message } from '@/types/message';

export const chatApi = {
  list: () => api.get<ChatSummary[]>('/chats'),
  getById: (id: string) => api.get<Chat>(`/chats/${id}`),
  create: (body: { memberIds: string[] }) => api.post<Chat>('/chats', body),
  getMessages: (chatId: string, cursor?: string) =>
    api.get<Message[]>(`/chats/${chatId}/messages${cursor ? `?cursor=${cursor}` : ''}`),
  sendMessage: (chatId: string, body: { text: string; replyTo?: string }) =>
    api.post<Message>(`/chats/${chatId}/messages`, body),
};
```

---

## Styling

### StyleSheet.create Pattern

```typescript
// Always use StyleSheet.create (not inline styles)
const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: '#0B141A',
  },
  header: {
    flexDirection: 'row',
    alignItems: 'center',
    paddingHorizontal: 16,
    paddingVertical: 12,
  },
  title: {
    fontSize: 18,
    fontWeight: '600',
    color: '#E9EDEF',
  },
});
```

### Color Palette (WhatsApp Dark Theme)

```typescript
// lib/constants.ts

export const COLORS = {
  background: '#0B141A',
  surface: '#1F2C34',
  surfaceLight: '#233138',
  primary: '#00A884',
  primaryDark: '#005C4B',
  textPrimary: '#E9EDEF',
  textSecondary: '#8696A0',
  textMuted: '#667781',
  border: '#233138',
  danger: '#EA4335',
  bubbleMine: '#005C4B',
  bubbleTheirs: '#1F2C34',
  unread: '#00A884',
} as const;
```

### Style Rules
- No inline styles (except truly dynamic values)
- Use `StyleSheet.create` for static styles
- Combine styles with array syntax: `style={[styles.base, isActive && styles.active]}`
- Keep color values in constants file
- Use `COLORS.xxx` instead of hardcoded hex

---

## i18n (Internationalization)

```typescript
// lib/i18n.ts

import i18n from 'i18next';
import { initReactI18next } from 'react-i18next';
import id from '@/i18n/id.json';
import en from '@/i18n/en.json';
import ar from '@/i18n/ar.json';

i18n.use(initReactI18next).init({
  resources: {
    id: { translation: id },
    en: { translation: en },
    ar: { translation: ar },
  },
  lng: 'id',
  fallbackLng: 'id',
  interpolation: { escapeValue: false },
});

export default i18n;

// Usage in components:
import { useTranslation } from 'react-i18next';

function ChatListScreen() {
  const { t } = useTranslation();
  return <Text>{t('chat.title')}</Text>;
}
```

---

## Types

```typescript
// types/chat.ts

export type ChatType = 'personal' | 'group';

export type Chat = {
  id: string;
  type: ChatType;
  name: string | null;
  icon: string | null;
  memberIds: string[];
  adminIds: string[];
  createdAt: string;
};

export type ChatSummary = {
  id: string;
  type: ChatType;
  name: string;
  icon: string | null;
  lastMessage: Message | null;
  unreadCount: number;
  updatedAt: string;
};

// types/document.ts

export type Document = {
  id: string;
  title: string;
  icon: string;
  cover: string | null;
  blocks: Block[];
  tags: string[];
  entities: string[];
  ownerId: string;
  locked: boolean;
  lockedAt: string | null;
  createdAt: string;
  updatedAt: string;
};

export type Block = {
  id: string;
  type: BlockType;
  content: string | null;
  checked: boolean | null;
  children: Block[] | null;
  rows: string[][] | null;
  columns: { name: string; type: string }[] | null;
  language: string | null;
  emoji: string | null;
  color: string | null;
};

export type BlockType =
  | 'paragraph'
  | 'heading1'
  | 'heading2'
  | 'heading3'
  | 'bullet-list'
  | 'numbered-list'
  | 'checklist'
  | 'table'
  | 'callout'
  | 'code'
  | 'toggle'
  | 'divider'
  | 'quote';
```

### Type Rules
- Use `type` for plain data shapes
- Use `interface` only when extending is needed
- No `enum` keyword — use union types instead: `type ChatType = 'personal' | 'group'`
- Suffix Props types: `MessageBubbleProps`
- Keep types close to where they're used; share via `types/` folder

---

## Performance Guidelines

1. **FlatList:** Always use `FlatList` for lists, never `ScrollView` with `.map()`
2. **Memoization:** Use `React.memo()` for expensive render components
3. **Callbacks:** Wrap event handlers with `useCallback`
4. **Images:** Use proper image sizing, cache with `react-native-fast-image`
5. **Re-renders:** Use Zustand selectors to prevent unnecessary re-renders
6. **Keyboard:** Handle keyboard avoidance properly for chat input
7. **Navigation:** Use lazy loading for screens not immediately visible
8. **Lists:** Use `keyExtractor`, `getItemLayout` for optimized scrolling

---

## Forbidden Practices

| Practice | Why | Alternative |
|----------|-----|------------|
| `any` type | No type safety | `unknown` or proper types |
| Inline styles | Performance hit | `StyleSheet.create` |
| `ScrollView` + `.map()` | Memory leak on long lists | `FlatList` |
| Class components | Outdated pattern | Functional components |
| `var` keyword | Scope issues | `const` / `let` |
| Side effects in render | Bugs, infinite loops | `useEffect` |
| Hardcoded strings | i18n issues | Translation keys |
| `console.log` in production | Performance | Remove or use proper logging |
| Direct store mutation | Unpredictable state | Zustand actions |
| Non-null assertion (`x!`) | Unsafe | Proper null checks |
