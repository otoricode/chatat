// Badge component — unread count, status indicators
import React from 'react';
import { View, Text, StyleSheet } from 'react-native';
import { colors, fontSize, fontFamily } from '@/theme';

type BadgeVariant = 'unread' | 'draft' | 'locked' | 'signature';

type BadgeProps = {
  count?: number;
  variant: BadgeVariant;
};

const VARIANT_COLORS: Record<BadgeVariant, string> = {
  unread: colors.green,
  draft: colors.textMuted,
  locked: colors.yellow,
  signature: colors.purple,
};

export function Badge({ count, variant }: BadgeProps) {
  const bg = VARIANT_COLORS[variant];
  const displayCount = count !== undefined && count > 99 ? '99+' : count?.toString();

  return (
    <View style={[styles.badge, { backgroundColor: bg }]}>
      {displayCount !== undefined && (
        <Text style={styles.text}>{displayCount}</Text>
      )}
    </View>
  );
}

const styles = StyleSheet.create({
  badge: {
    minWidth: 20,
    height: 20,
    borderRadius: 10,
    justifyContent: 'center',
    alignItems: 'center',
    paddingHorizontal: 6,
  },
  text: {
    fontFamily: fontFamily.uiSemiBold,
    fontSize: fontSize.xs,
    color: colors.background,
  },
});
