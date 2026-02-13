// DateSeparator — date divider between messages of different days
import React from 'react';
import { View, Text, StyleSheet } from 'react-native';
import { useTranslation } from 'react-i18next';
import { colors, fontSize, fontFamily, spacing } from '@/theme';
import { formatDateSeparator } from '@/lib/timeFormat';

type Props = {
  dateStr: string;
};

function DateSeparatorInner({ dateStr }: Props) {
  const { t } = useTranslation();
  return (
    <View style={styles.container}>
      <View style={styles.badge}>
        <Text style={styles.text}>{formatDateSeparator(dateStr, t)}</Text>
      </View>
    </View>
  );
}

const styles = StyleSheet.create({
  container: {
    alignItems: 'center',
    paddingVertical: spacing.sm,
  },
  badge: {
    backgroundColor: colors.surface2,
    paddingHorizontal: spacing.md,
    paddingVertical: spacing.xs,
    borderRadius: 8,
  },
  text: {
    fontFamily: fontFamily.uiMedium,
    fontSize: fontSize.xs,
    color: colors.textMuted,
  },
});

export const DateSeparator = React.memo(DateSeparatorInner);
