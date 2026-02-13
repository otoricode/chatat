// Subtle sync indicator — shows ActivityIndicator when syncing
import React from 'react';
import { ActivityIndicator } from 'react-native';
import { useSyncStore } from '@/stores/syncStore';
import { colors } from '@/theme';

export const SyncIndicator: React.FC = () => {
  const isSyncing = useSyncStore((s) => s.isSyncing);

  if (!isSyncing) return null;

  return <ActivityIndicator size="small" color={colors.green} />;
};
