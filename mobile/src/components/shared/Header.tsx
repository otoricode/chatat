// Header — app header with title, search, and profile
// Based on spesifikasi-chatat.md section 7.2
import React from 'react';
import { View, Text, StyleSheet, Pressable } from 'react-native';
import { colors, fontSize, spacing, fontFamily } from '@/theme';

type HeaderProps = {
  title: string;
  onSearchPress?: () => void;
  onProfilePress?: () => void;
};

export function Header({ title, onSearchPress, onProfilePress }: HeaderProps) {
  return (
    <View style={styles.container}>
      <Text style={styles.title}>{title}</Text>
      <View style={styles.actions}>
        {onSearchPress && (
          <Pressable onPress={onSearchPress} style={styles.iconButton}>
            <Text style={styles.icon}>🔍</Text>
          </Pressable>
        )}
        {onProfilePress && (
          <Pressable onPress={onProfilePress} style={styles.iconButton}>
            <Text style={styles.icon}>👤</Text>
          </Pressable>
        )}
      </View>
    </View>
  );
}

const styles = StyleSheet.create({
  container: {
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'space-between',
    paddingHorizontal: spacing.lg,
    paddingVertical: spacing.md,
    backgroundColor: colors.headerBackground,
  },
  title: {
    fontFamily: fontFamily.uiBold,
    fontSize: fontSize.xl,
    color: colors.green,
  },
  actions: {
    flexDirection: 'row',
    gap: spacing.xs,
  },
  iconButton: {
    padding: spacing.sm,
    borderRadius: 20,
  },
  icon: {
    fontSize: 20,
  },
});
