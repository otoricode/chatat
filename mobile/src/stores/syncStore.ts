// Zustand store for sync status
import { create } from 'zustand';
import { syncEngine, type SyncStatus } from '@/services/SyncEngine';

interface SyncState extends SyncStatus {
  startSync: () => Promise<void>;
  syncMessages: (chatId: string) => Promise<void>;
}

export const useSyncStore = create<SyncState>((set) => {
  // Subscribe to SyncEngine status updates
  syncEngine.subscribe((status) => {
    set({
      isSyncing: status.isSyncing,
      lastSyncedAt: status.lastSyncedAt,
      error: status.error,
    });
  });

  return {
    isSyncing: false,
    lastSyncedAt: 0,
    error: null,

    startSync: async () => {
      await syncEngine.fullSync();
    },

    syncMessages: async (chatId: string) => {
      await syncEngine.syncMessages(chatId);
    },
  };
});
