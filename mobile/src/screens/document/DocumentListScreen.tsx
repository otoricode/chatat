// Document list screen — main document tab
import React from 'react';
import { View, StyleSheet } from 'react-native';
import { SafeAreaView } from 'react-native-safe-area-context';
import type { NativeStackScreenProps } from '@react-navigation/native-stack';
import type { DocumentStackParamList } from '@/navigation/types';
import { Header } from '@/components/shared/Header';
import { FAB } from '@/components/shared/FAB';
import { EmptyState } from '@/components/shared/EmptyState';
import { colors, spacing } from '@/theme';

type Props = NativeStackScreenProps<DocumentStackParamList, 'DocumentList'>;

export function DocumentListScreen({ navigation }: Props) {
  const handleNewDocument = () => {
    navigation.navigate('DocumentEditor', {});
  };

  return (
    <SafeAreaView style={styles.container} edges={['top']}>
      <Header title="Dokumen" />
      <View style={styles.content}>
        <EmptyState
          emoji="📄"
          title="Belum ada dokumen"
          description="Buat dokumen baru untuk mulai berkolaborasi"
        />
      </View>
      <FAB onPress={handleNewDocument} />
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
