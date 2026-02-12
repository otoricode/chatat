// Document viewer screen — for locked documents (placeholder)
import React from 'react';
import { View, Text, StyleSheet } from 'react-native';
import { SafeAreaView } from 'react-native-safe-area-context';
import type { NativeStackScreenProps } from '@react-navigation/native-stack';
import type { DocumentStackParamList } from '@/navigation/types';
import { colors, fontSize, fontFamily } from '@/theme';

type Props = NativeStackScreenProps<DocumentStackParamList, 'DocumentViewer'>;

export function DocumentViewerScreen({ route }: Props) {
  const { documentId } = route.params;

  return (
    <SafeAreaView style={styles.container}>
      <View style={styles.content}>
        <Text style={styles.lockIcon}>🔒</Text>
        <Text style={styles.text}>Dokumen Terkunci</Text>
        <Text style={styles.subtext}>{documentId}</Text>
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
    justifyContent: 'center',
    alignItems: 'center',
  },
  lockIcon: {
    fontSize: 48,
    marginBottom: 12,
  },
  text: {
    fontFamily: fontFamily.uiSemiBold,
    fontSize: fontSize.lg,
    color: colors.textPrimary,
  },
  subtext: {
    fontFamily: fontFamily.ui,
    fontSize: fontSize.sm,
    color: colors.textMuted,
    marginTop: 4,
  },
});
