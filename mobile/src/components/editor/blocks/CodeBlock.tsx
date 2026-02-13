// CodeBlock — monospace code editor
import React, { useRef, useEffect } from 'react';
import { View, TextInput, Text, Pressable, StyleSheet } from 'react-native';
import * as Clipboard from 'expo-clipboard';
import { useTranslation } from 'react-i18next';
import { colors, fontFamily, fontSize, spacing } from '@/theme';
import type { BlockProps } from '../types';

export const CodeBlock = React.memo(function CodeBlock({
  block,
  isActive,
  readOnly,
  onChange,
  onFocus,
}: BlockProps) {
  const { t } = useTranslation();
  const inputRef = useRef<TextInput>(null);

  useEffect(() => {
    if (isActive && inputRef.current) {
      inputRef.current.focus();
    }
  }, [isActive]);

  const handleCopy = async () => {
    if (block.content) {
      await Clipboard.setStringAsync(block.content);
    }
  };

  return (
    <View style={styles.container}>
      <View style={styles.header}>
        <Text style={styles.language}>{block.language || 'text'}</Text>
        <Pressable onPress={handleCopy} style={styles.copyBtn}>
          <Text style={styles.copyText}>{t('common.copy')}</Text>
        </Pressable>
      </View>
      <TextInput
        ref={inputRef}
        style={styles.input}
        value={block.content}
        onChangeText={(text) => onChange({ content: text })}
        onFocus={onFocus}
        multiline
        editable={!readOnly}
        placeholder={t('editor.codePlaceholder')}
        placeholderTextColor={colors.textMuted}
        autoCapitalize="none"
        autoCorrect={false}
        spellCheck={false}
      />
    </View>
  );
});

const styles = StyleSheet.create({
  container: {
    marginHorizontal: spacing.lg,
    marginVertical: spacing.xs,
    backgroundColor: colors.surface,
    borderRadius: 8,
    borderWidth: 1,
    borderColor: colors.border,
    overflow: 'hidden',
  },
  header: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    paddingHorizontal: spacing.md,
    paddingVertical: spacing.xs,
    backgroundColor: colors.surface2,
    borderBottomWidth: 1,
    borderBottomColor: colors.border,
  },
  language: {
    fontFamily: fontFamily.code,
    fontSize: fontSize.xs,
    color: colors.textMuted,
  },
  copyBtn: {
    paddingHorizontal: spacing.sm,
    paddingVertical: 2,
  },
  copyText: {
    fontFamily: fontFamily.ui,
    fontSize: fontSize.xs,
    color: colors.blue,
  },
  input: {
    fontFamily: fontFamily.code,
    fontSize: fontSize.sm,
    lineHeight: fontSize.sm * 1.6,
    color: colors.textPrimary,
    padding: spacing.md,
    minHeight: 60,
  },
});
