// Topic list screen (placeholder)
import React from 'react';
import { View, StyleSheet } from 'react-native';
import { SafeAreaView } from 'react-native-safe-area-context';
import type { NativeStackScreenProps } from '@react-navigation/native-stack';
import type { ChatStackParamList } from '@/navigation/types';
import { EmptyState } from '@/components/shared/EmptyState';
import { colors, spacing } from '@/theme';

type Props = NativeStackScreenProps<ChatStackParamList, 'TopicList'>;

export function TopicListScreen(_props: Props) {
  return (
    <SafeAreaView style={styles.container}>
      <View style={styles.content}>
        <EmptyState
          emoji="📌"
          title="Belum ada topik"
          description="Buat topik untuk diskusi terfokus"
        />
      </View>
    </SafeAreaView>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: colors.background,
  },
  content: {
    flex: 1,
    paddingHorizontal: spacing.lg,
  },
});
