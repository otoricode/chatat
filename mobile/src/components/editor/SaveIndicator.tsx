// SaveIndicator — shows save status
import React from 'react';
import { View, Text, ActivityIndicator, StyleSheet } from 'react-native';
import { useEditorStore } from '@/stores/editorStore';
import { colors, fontFamily, fontSize, spacing } from '@/theme';

export function SaveIndicator() {
  const saveStatus = useEditorStore((s) => s.saveStatus);

  if (saveStatus === 'idle') return null;

  return (
    <View style={styles.container}>
      {saveStatus === 'saving' && (
        <>
          <ActivityIndicator size="small" color={colors.textMuted} />
          <Text style={styles.text}>Menyimpan...</Text>
        </>
      )}
      {saveStatus === 'saved' && (
        <>
          <Text style={styles.check}>✓</Text>
          <Text style={styles.text}>Tersimpan</Text>
        </>
      )}
      {saveStatus === 'error' && (
        <>
          <Text style={styles.errorIcon}>⚠</Text>
          <Text style={[styles.text, styles.errorText]}>Gagal menyimpan</Text>
        </>
      )}
    </View>
  );
}

const styles = StyleSheet.create({
  container: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: spacing.xs,
  },
  text: {
    fontFamily: fontFamily.ui,
    fontSize: fontSize.xs,
    color: colors.textMuted,
  },
  check: {
    fontSize: 12,
    color: colors.green,
  },
  errorIcon: {
    fontSize: 12,
    color: colors.yellow,
  },
  errorText: {
    color: colors.yellow,
  },
});
