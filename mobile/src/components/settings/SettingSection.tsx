// SettingSection — grouping wrapper for settings rows
import React from 'react';
import { View, Text, StyleSheet } from 'react-native';
import { colors, fontFamily, fontSize, spacing } from '@/theme';

type SettingSectionProps = {
  title: string;
  children: React.ReactNode;
};

export function SettingSection({ title, children }: SettingSectionProps) {
  return (
    <View style={styles.container}>
      <Text style={styles.title}>{title}</Text>
      <View style={styles.card}>{children}</View>
    </View>
  );
}

const styles = StyleSheet.create({
  container: {
    marginBottom: spacing.lg,
  },
  title: {
    fontFamily: fontFamily.uiSemiBold,
    fontSize: fontSize.sm,
    color: colors.textMuted,
    textTransform: 'uppercase',
    letterSpacing: 0.5,
    marginBottom: spacing.sm,
    paddingHorizontal: spacing.xs,
  },
  card: {
    backgroundColor: colors.surface,
    borderRadius: 12,
    overflow: 'hidden',
  },
});
