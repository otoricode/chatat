// Settings store — manages user preferences
// Persisted with AsyncStorage
import { create } from 'zustand';
import { persist, createJSONStorage } from 'zustand/middleware';
import AsyncStorage from '@react-native-async-storage/async-storage';

type NotificationPrefs = {
  showPreview: boolean;
  soundEnabled: boolean;
  vibrationEnabled: boolean;
  groupAlerts: boolean;
};

type AutoDownloadPrefs = {
  wifiAll: boolean;
  cellularUnder5MB: boolean;
  cellularAsk: boolean;
};

type SettingsState = {
  // Notification preferences
  notifications: NotificationPrefs;
  // Auto-download preferences
  autoDownload: AutoDownloadPrefs;

  // Actions
  updateNotifications: (updates: Partial<NotificationPrefs>) => void;
  updateAutoDownload: (updates: Partial<AutoDownloadPrefs>) => void;
  reset: () => void;
};

const defaultNotifications: NotificationPrefs = {
  showPreview: true,
  soundEnabled: true,
  vibrationEnabled: true,
  groupAlerts: true,
};

const defaultAutoDownload: AutoDownloadPrefs = {
  wifiAll: true,
  cellularUnder5MB: true,
  cellularAsk: true,
};

export const useSettingsStore = create<SettingsState>()(
  persist(
    (set) => ({
      notifications: { ...defaultNotifications },
      autoDownload: { ...defaultAutoDownload },

      updateNotifications: (updates) =>
        set((state) => ({
          notifications: { ...state.notifications, ...updates },
        })),

      updateAutoDownload: (updates) =>
        set((state) => ({
          autoDownload: { ...state.autoDownload, ...updates },
        })),

      reset: () =>
        set({
          notifications: { ...defaultNotifications },
          autoDownload: { ...defaultAutoDownload },
        }),
    }),
    {
      name: 'settings-storage',
      storage: createJSONStorage(() => AsyncStorage),
    },
  ),
);
