// ScreenContainer — SafeAreaView wrapper for consistent screen layout
import React from 'react';
import { StyleSheet, type ViewStyle } from 'react-native';
import { SafeAreaView, type Edge } from 'react-native-safe-area-context';
import { colors } from '@/theme';

type ScreenContainerProps = {
  children: React.ReactNode;
  edges?: Edge[];
  style?: ViewStyle;
};

export function ScreenContainer({
  children,
  edges = ['top'],
  style,
}: ScreenContainerProps) {
  return (
    <SafeAreaView style={[styles.container, style]} edges={edges}>
      {children}
    </SafeAreaView>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: colors.background,
  },
});
