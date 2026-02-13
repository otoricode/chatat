// Hook for loading data from local DB, with server fallback when online.
import { useState, useEffect, useCallback } from 'react';
import { useNetworkStore } from '@/stores/networkStore';

type FetchFn<T> = () => Promise<T>;

interface UseOfflineDataOptions<T> {
  /** Fetch from server API */
  serverFetch: FetchFn<T>;
  /** Fetch from local SQLite */
  localFetch: FetchFn<T>;
  /** Whether to auto-fetch on mount */
  autoFetch?: boolean;
}

interface UseOfflineDataResult<T> {
  data: T | null;
  loading: boolean;
  error: string | null;
  isOffline: boolean;
  refresh: () => Promise<void>;
}

/**
 * Loads data from server when online, falls back to local DB when offline.
 * Always saves server data locally via the sync engine.
 */
export function useOfflineData<T>({
  serverFetch,
  localFetch,
  autoFetch = true,
}: UseOfflineDataOptions<T>): UseOfflineDataResult<T> {
  const [data, setData] = useState<T | null>(null);
  const [loading, setLoading] = useState(autoFetch);
  const [error, setError] = useState<string | null>(null);
  const isConnected = useNetworkStore((s) => s.isConnected);

  const fetchData = useCallback(async () => {
    setLoading(true);
    setError(null);

    try {
      if (isConnected) {
        // Try server first
        try {
          const serverData = await serverFetch();
          setData(serverData);
          setLoading(false);
          return;
        } catch {
          // Server failed, fall back to local
        }
      }

      // Offline or server failed: use local data
      const localData = await localFetch();
      setData(localData);
    } catch (err) {
      const msg = err instanceof Error ? err.message : 'Failed to load data';
      setError(msg);
    } finally {
      setLoading(false);
    }
  }, [isConnected, serverFetch, localFetch]);

  useEffect(() => {
    if (autoFetch) {
      fetchData();
    }
  }, [autoFetch, fetchData]);

  return {
    data,
    loading,
    error,
    isOffline: !isConnected,
    refresh: fetchData,
  };
}
