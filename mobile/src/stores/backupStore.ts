// Backup store — manages backup state and operations
import { create } from 'zustand';
import { backupApi } from '@/services/api/backup';
import type { BackupRecord } from '@/services/api/backup';
import {
  performBackup,
  performRestore,
  listCloudBackups,
  getPlatform,
} from '@/services/backup/CloudBackupService';
import type { BackupFile, BackupProgress } from '@/services/backup/CloudBackupService';

type BackupState = {
  // State
  latestBackup: BackupRecord | null;
  history: BackupRecord[];
  cloudBackups: BackupFile[];
  isLoading: boolean;
  isBackingUp: boolean;
  isRestoring: boolean;
  progress: BackupProgress | null;
  error: string | null;
  platform: string;

  // Actions
  fetchLatest: () => Promise<void>;
  fetchHistory: () => Promise<void>;
  fetchCloudBackups: () => Promise<void>;
  startBackup: () => Promise<void>;
  startRestore: () => Promise<void>;
  clearError: () => void;
};

export const useBackupStore = create<BackupState>()((set) => ({
  latestBackup: null,
  history: [],
  cloudBackups: [],
  isLoading: false,
  isBackingUp: false,
  isRestoring: false,
  progress: null,
  error: null,
  platform: getPlatform(),

  fetchLatest: async () => {
    try {
      const res = await backupApi.getLatest();
      const apiData = res.data as unknown as { success: boolean; data: BackupRecord | null };
      set({ latestBackup: apiData.data ?? null });
    } catch (err) {
      const msg = err instanceof Error ? err.message : 'Failed to fetch latest backup';
      set({ error: msg });
    }
  },

  fetchHistory: async () => {
    set({ isLoading: true, error: null });
    try {
      const res = await backupApi.getHistory();
      const apiData = res.data as unknown as { success: boolean; data: BackupRecord[] };
      set({ history: apiData.data ?? [], isLoading: false });
    } catch (err) {
      const msg = err instanceof Error ? err.message : 'Failed to fetch backup history';
      set({ error: msg, isLoading: false });
    }
  },

  fetchCloudBackups: async () => {
    try {
      const backups = await listCloudBackups();
      set({ cloudBackups: backups });
    } catch {
      // Cloud list can fail silently without signed in
    }
  },

  startBackup: async () => {
    set({ isBackingUp: true, progress: null, error: null });
    try {
      await performBackup((progress) => {
        set({ progress });
      });
      // Refresh after backup
      const res = await backupApi.getLatest();
      const apiData = res.data as unknown as { success: boolean; data: BackupRecord | null };
      set({ latestBackup: apiData.data ?? null, isBackingUp: false, progress: null });
    } catch (err) {
      const msg = err instanceof Error ? err.message : 'Backup failed';
      set({ error: msg, isBackingUp: false, progress: null });
    }
  },

  startRestore: async () => {
    set({ isRestoring: true, progress: null, error: null });
    try {
      await performRestore((progress) => {
        set({ progress });
      });
      set({ isRestoring: false, progress: null });
    } catch (err) {
      const msg = err instanceof Error ? err.message : 'Restore failed';
      set({ error: msg, isRestoring: false, progress: null });
    }
  },

  clearError: () => set({ error: null }),
}));
