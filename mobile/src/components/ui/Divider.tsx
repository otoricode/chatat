// Divider component — horizontal separator
import React from 'react';
import { View, StyleSheet } from 'react-native';
import { colors, spacing } from '@/theme';

type DividerProps = {
  marginVertical?: number;
};

export function Divider({ marginVertical = spacing.md }: DividerProps) {
  return <View style={[styles.divider, { marginVertical }]} />;
}

const styles = StyleSheet.create({
  divider: {
    height: 0.5,
    backgroundColor: colors.border,
  },
});
