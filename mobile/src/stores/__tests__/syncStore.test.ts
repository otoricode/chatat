// @ts-nocheck
jest.mock('@/services/SyncEngine', () => {
  const listeners: ((status: any) => void)[] = [];
  return {
    syncEngine: {
      subscribe: jest.fn((listener: any) => {
        listeners.push(listener);
        return () => {
          const idx = listeners.indexOf(listener);
          if (idx >= 0) listeners.splice(idx, 1);
        };
      }),
      fullSync: jest.fn(),
      syncMessages: jest.fn(),
      _emit: (status: any) => {
        for (const l of listeners) l(status);
      },
    },
  };
});

import { useSyncStore } from '../syncStore';
import { syncEngine } from '@/services/SyncEngine';

const mockEngine = syncEngine as any;

beforeEach(() => {
  useSyncStore.setState({
    isSyncing: false,
    lastSyncedAt: 0,
    error: null,
  });
  jest.clearAllMocks();
});

describe('syncStore', () => {
  it('starts not syncing', () => {
    const s = useSyncStore.getState();
    expect(s.isSyncing).toBe(false);
    expect(s.lastSyncedAt).toBe(0);
    expect(s.error).toBeNull();
  });

  it('startSync calls fullSync', async () => {
    mockEngine.fullSync.mockResolvedValue(undefined);

    await useSyncStore.getState().startSync();

    expect(mockEngine.fullSync).toHaveBeenCalled();
  });

  it('syncMessages calls engine syncMessages', async () => {
    mockEngine.syncMessages.mockResolvedValue(undefined);

    await useSyncStore.getState().syncMessages('chat1');

    expect(mockEngine.syncMessages).toHaveBeenCalledWith('chat1');
  });

  it('receives status updates from engine', () => {
    mockEngine._emit({
      isSyncing: true,
      lastSyncedAt: 12345,
      error: null,
    });

    const s = useSyncStore.getState();
    expect(s.isSyncing).toBe(true);
    expect(s.lastSyncedAt).toBe(12345);
  });

  it('receives error from engine', () => {
    mockEngine._emit({
      isSyncing: false,
      lastSyncedAt: 0,
      error: 'Sync failed',
    });

    expect(useSyncStore.getState().error).toBe('Sync failed');
  });
});
