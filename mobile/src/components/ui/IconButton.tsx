// IconButton component — icon-only pressable
import React from 'react';
import { Pressable, Text, StyleSheet, type ViewStyle } from 'react-native';
import { colors, spacing } from '@/theme';

type IconButtonProps = {
  icon: string;
  onPress: () => void;
  size?: number;
  style?: ViewStyle;
};

export function IconButton({ icon, onPress, size = 24, style }: IconButtonProps) {
  return (
    <Pressable
      style={({ pressed }) => [styles.button, pressed && styles.pressed, style]}
      onPress={onPress}
    >
      <Text style={{ fontSize: size }}>{icon}</Text>
    </Pressable>
  );
}

const styles = StyleSheet.create({
  button: {
    padding: spacing.sm,
    borderRadius: 20,
    justifyContent: 'center',
    alignItems: 'center',
  },
  pressed: {
    backgroundColor: colors.surface2,
  },
});
