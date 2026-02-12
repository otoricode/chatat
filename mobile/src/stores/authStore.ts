// Auth store — manages authentication state
// Persisted with AsyncStorage

import { create } from 'zustand';
import { persist, createJSONStorage } from 'zustand/middleware';
import AsyncStorage from '@react-native-async-storage/async-storage';

type User = {
  id: string;
  name: string;
  phone: string;
  avatar: string;
  status: string;
};

type TokenPair = {
  accessToken: string;
  refreshToken: string;
};

type AuthState = {
  isAuthenticated: boolean;
  accessToken: string | null;
  refreshToken: string | null;
  user: User | null;
  isNewUser: boolean;

  login: (tokens: TokenPair, user: User, isNew: boolean) => void;
  logout: () => void;
  setUser: (user: User) => void;
  updateProfile: (updates: Partial<User>) => void;
  setTokens: (tokens: TokenPair) => void;
};

export const useAuthStore = create<AuthState>()(
  persist(
    (set) => ({
      isAuthenticated: false,
      accessToken: null,
      refreshToken: null,
      user: null,
      isNewUser: false,

      login: (tokens, user, isNew) =>
        set({
          isAuthenticated: true,
          accessToken: tokens.accessToken,
          refreshToken: tokens.refreshToken,
          user,
          isNewUser: isNew,
        }),

      logout: () =>
        set({
          isAuthenticated: false,
          accessToken: null,
          refreshToken: null,
          user: null,
          isNewUser: false,
        }),

      setUser: (user) => set({ user }),

      updateProfile: (updates) =>
        set((state) => ({
          user: state.user ? { ...state.user, ...updates } : null,
        })),

      setTokens: (tokens) =>
        set({
          accessToken: tokens.accessToken,
          refreshToken: tokens.refreshToken,
        }),
    }),
    {
      name: 'auth-storage',
      storage: createJSONStorage(() => AsyncStorage),
    },
  ),
);
