// CalloutBlock — emoji + colored background + text
import React, { useRef, useEffect } from 'react';
import { View, TextInput, Text, StyleSheet } from 'react-native';
import { useTranslation } from 'react-i18next';
import { colors, fontFamily, fontSize, spacing } from '@/theme';
import type { BlockProps } from '../types';

const CALLOUT_COLORS: Record<string, string> = {
  blue: 'rgba(96,165,250,0.15)',
  green: 'rgba(110,231,183,0.15)',
  yellow: 'rgba(251,191,36,0.15)',
  red: 'rgba(248,113,113,0.15)',
  default: 'rgba(110,231,183,0.1)',
};

export const CalloutBlock = React.memo(function CalloutBlock({
  block,
  isActive,
  readOnly,
  onChange,
  onFocus,
}: BlockProps) {
  const { t } = useTranslation();
  const inputRef = useRef<TextInput>(null);
  const bgColor = CALLOUT_COLORS[block.color ?? 'default'] ?? CALLOUT_COLORS.default;

  useEffect(() => {
    if (isActive && inputRef.current) {
      inputRef.current.focus();
    }
  }, [isActive]);

  return (
    <View style={[styles.container, { backgroundColor: bgColor }]}>
      <Text style={styles.emoji}>{block.emoji || '💡'}</Text>
      <TextInput
        ref={inputRef}
        style={styles.input}
        value={block.content}
        onChangeText={(text) => onChange({ content: text })}
        onFocus={onFocus}
        multiline
        editable={!readOnly}
        placeholder={t('editor.calloutPlaceholder')}
        placeholderTextColor={colors.textMuted}
        blurOnSubmit={false}
      />
    </View>
  );
});

const styles = StyleSheet.create({
  container: {
    flexDirection: 'row',
    alignItems: 'flex-start',
    marginHorizontal: spacing.lg,
    marginVertical: spacing.xs,
    padding: spacing.md,
    borderRadius: 8,
  },
  emoji: {
    fontSize: 20,
    marginRight: spacing.sm,
    marginTop: 2,
  },
  input: {
    flex: 1,
    fontFamily: fontFamily.document,
    fontSize: fontSize.md,
    lineHeight: fontSize.md * 1.5,
    color: colors.textPrimary,
    minHeight: 24,
    padding: 0,
  },
});
