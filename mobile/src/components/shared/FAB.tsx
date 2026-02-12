// FAB — Floating Action Button
// Based on spesifikasi-chatat.md section 7.3
import React from 'react';
import { Pressable, Text, StyleSheet } from 'react-native';
import { colors } from '@/theme';

type FABProps = {
  onPress: () => void;
  icon?: string;
};

export function FAB({ onPress, icon = '+' }: FABProps) {
  return (
    <Pressable
      style={({ pressed }) => [styles.fab, pressed && styles.pressed]}
      onPress={onPress}
    >
      <Text style={styles.icon}>{icon}</Text>
    </Pressable>
  );
}

const styles = StyleSheet.create({
  fab: {
    position: 'absolute',
    bottom: 24,
    right: 20,
    width: 56,
    height: 56,
    borderRadius: 28,
    backgroundColor: colors.green,
    justifyContent: 'center',
    alignItems: 'center',
    elevation: 6,
    shadowColor: colors.black,
    shadowOffset: { width: 0, height: 3 },
    shadowOpacity: 0.3,
    shadowRadius: 4,
  },
  pressed: {
    opacity: 0.85,
    transform: [{ scale: 0.95 }],
  },
  icon: {
    fontSize: 28,
    color: colors.background,
    fontWeight: '300',
    marginTop: -2,
  },
});
