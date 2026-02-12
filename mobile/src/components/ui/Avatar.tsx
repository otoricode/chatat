// Avatar component — emoji avatar with optional online indicator
import React from 'react';
import { View, Text, StyleSheet } from 'react-native';
import { colors } from '@/theme';

type AvatarProps = {
  emoji: string;
  size?: 'sm' | 'md' | 'lg';
  online?: boolean;
};

const SIZES = {
  sm: { container: 32, font: 16, dot: 8 },
  md: { container: 44, font: 24, dot: 10 },
  lg: { container: 56, font: 32, dot: 12 },
} as const;

export function Avatar({ emoji, size = 'md', online }: AvatarProps) {
  const s = SIZES[size];

  return (
    <View
      style={[
        styles.container,
        {
          width: s.container,
          height: s.container,
          borderRadius: s.container / 2,
        },
      ]}
    >
      <Text style={{ fontSize: s.font }}>{emoji}</Text>
      {online !== undefined && (
        <View
          style={[
            styles.dot,
            {
              width: s.dot,
              height: s.dot,
              borderRadius: s.dot / 2,
              backgroundColor: online ? colors.online : colors.offline,
            },
          ]}
        />
      )}
    </View>
  );
}

const styles = StyleSheet.create({
  container: {
    backgroundColor: colors.surface,
    justifyContent: 'center',
    alignItems: 'center',
  },
  dot: {
    position: 'absolute',
    bottom: 0,
    right: 0,
    borderWidth: 2,
    borderColor: colors.background,
  },
});
