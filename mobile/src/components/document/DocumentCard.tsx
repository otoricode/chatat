// Document card for displaying document links in chat messages
import React from 'react';
import { View, Text, Pressable, StyleSheet } from 'react-native';
import { colors, fontSize, fontFamily, spacing } from '@/theme';
import { useTranslation } from 'react-i18next';
import { LockStatusBadge } from './LockStatusBadge';

interface DocumentCardProps {
  title: string;
  icon: string;
  locked: boolean;
  lockedBy?: string | null;
  updatedAt?: string;
  onPress: () => void;
}

export function DocumentCard({
  title,
  icon,
  locked,
  lockedBy,
  updatedAt,
  onPress,
}: DocumentCardProps) {
  const { t } = useTranslation();
  return (
    <Pressable style={styles.card} onPress={onPress}>
      <View style={styles.iconContainer}>
        <Text style={styles.icon}>{icon || '📄'}</Text>
      </View>
      <View style={styles.content}>
        <Text style={styles.title} numberOfLines={2}>
          {title || t('document.untitled')}
        </Text>
        <View style={styles.meta}>
          <LockStatusBadge locked={locked} lockedBy={lockedBy} compact />
          {updatedAt && (
            <Text style={styles.date}>{updatedAt}</Text>
          )}
        </View>
      </View>
      <Text style={styles.arrow}>›</Text>
    </Pressable>
  );
}

const styles = StyleSheet.create({
  card: {
    flexDirection: 'row',
    alignItems: 'center',
    backgroundColor: colors.surface2,
    borderRadius: 12,
    padding: spacing.md,
    gap: spacing.md,
    borderWidth: 1,
    borderColor: colors.border,
    maxWidth: 280,
  },
  iconContainer: {
    width: 40,
    height: 40,
    borderRadius: 8,
    backgroundColor: colors.surface,
    justifyContent: 'center',
    alignItems: 'center',
  },
  icon: {
    fontSize: 22,
  },
  content: {
    flex: 1,
  },
  title: {
    fontFamily: fontFamily.uiMedium,
    fontSize: fontSize.sm,
    color: colors.textPrimary,
    lineHeight: 18,
  },
  meta: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: spacing.xs,
    marginTop: spacing.xs,
  },
  date: {
    fontFamily: fontFamily.ui,
    fontSize: fontSize.xs,
    color: colors.textMuted,
  },
  arrow: {
    fontSize: fontSize.xl,
    color: colors.textMuted,
  },
});
