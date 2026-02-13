// Lock status badge component for documents
import React from 'react';
import { View, Text, StyleSheet } from 'react-native';
import { colors, fontSize, fontFamily, spacing } from '@/theme';
import { useTranslation } from 'react-i18next';

export type LockStatus = 'draft' | 'locked_manual' | 'pending_signatures' | 'locked_signed';

interface LockStatusBadgeProps {
  locked: boolean;
  lockedBy?: string | null;
  requireSigs?: boolean;
  compact?: boolean;
}

export function getLockStatus(locked: boolean, lockedBy?: string | null, requireSigs?: boolean): LockStatus {
  if (!locked) return 'draft';
  if (lockedBy === 'signatures') return 'locked_signed';
  if (requireSigs && !locked) return 'pending_signatures';
  return 'locked_manual';
}

const STATUS_CONFIG: Record<LockStatus, { labelKey: string; icon: string; color: string; bg: string }> = {
  draft: {
    labelKey: 'document.draft',
    icon: '📝',
    color: colors.textMuted,
    bg: 'transparent',
  },
  locked_manual: {
    labelKey: 'document.locked',
    icon: '🔒',
    color: colors.yellow,
    bg: 'rgba(251, 191, 36, 0.15)',
  },
  pending_signatures: {
    labelKey: 'document.pendingSignatures',
    icon: '✍️',
    color: colors.blue,
    bg: 'rgba(96, 165, 250, 0.15)',
  },
  locked_signed: {
    labelKey: 'document.signed',
    icon: '✅',
    color: colors.green,
    bg: 'rgba(110, 231, 183, 0.15)',
  },
};

export function LockStatusBadge({ locked, lockedBy, requireSigs, compact }: LockStatusBadgeProps) {
  const { t } = useTranslation();
  const status = getLockStatus(locked, lockedBy, requireSigs);

  if (status === 'draft' && compact) return null;

  const config = STATUS_CONFIG[status];

  if (compact) {
    return <Text style={styles.compactIcon}>{config.icon}</Text>;
  }

  return (
    <View style={[styles.badge, { backgroundColor: config.bg }]}>
      <Text style={styles.icon}>{config.icon}</Text>
      <Text style={[styles.label, { color: config.color }]}>{t(config.labelKey)}</Text>
    </View>
  );
}

const styles = StyleSheet.create({
  badge: {
    flexDirection: 'row',
    alignItems: 'center',
    paddingHorizontal: spacing.sm,
    paddingVertical: spacing.xs,
    borderRadius: 6,
    gap: spacing.xs,
  },
  icon: {
    fontSize: 14,
  },
  compactIcon: {
    fontSize: 18,
  },
  label: {
    fontFamily: fontFamily.uiMedium,
    fontSize: fontSize.xs,
  },
});
