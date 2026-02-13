// DividerBlock — horizontal line separator
import React from 'react';
import { View, StyleSheet } from 'react-native';
import { colors, spacing } from '@/theme';

export const DividerBlock = React.memo(function DividerBlock() {
  return (
    <View style={styles.container}>
      <View style={styles.line} />
    </View>
  );
});

const styles = StyleSheet.create({
  container: {
    paddingVertical: spacing.md,
    paddingHorizontal: spacing.lg,
  },
  line: {
    height: 1,
    backgroundColor: colors.border,
  },
});
