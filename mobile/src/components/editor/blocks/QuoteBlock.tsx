// QuoteBlock — left-bordered quote
import React, { useRef, useEffect } from 'react';
import { View, TextInput, StyleSheet } from 'react-native';
import { useTranslation } from 'react-i18next';
import { colors, fontFamily, fontSize, spacing } from '@/theme';
import type { BlockProps } from '../types';

export const QuoteBlock = React.memo(function QuoteBlock({
  block,
  isActive,
  readOnly,
  onChange,
  onFocus,
  onSubmit,
  onBackspace,
  onSlashTrigger,
}: BlockProps) {
  const { t } = useTranslation();
  const inputRef = useRef<TextInput>(null);

  useEffect(() => {
    if (isActive && inputRef.current) {
      inputRef.current.focus();
    }
  }, [isActive]);

  const handleChangeText = (text: string) => {
    if (text === '/') {
      onSlashTrigger();
      return;
    }
    onChange({ content: text });
  };

  const handleKeyPress = (e: { nativeEvent: { key: string } }) => {
    if (e.nativeEvent.key === 'Backspace' && block.content === '') {
      onBackspace();
    }
  };

  return (
    <View style={styles.container}>
      <View style={styles.border} />
      <TextInput
        ref={inputRef}
        style={styles.input}
        value={block.content}
        onChangeText={handleChangeText}
        onFocus={onFocus}
        onSubmitEditing={onSubmit}
        onKeyPress={handleKeyPress}
        multiline
        editable={!readOnly}
        placeholder={t('editor.quotePlaceholder')}
        placeholderTextColor={colors.textMuted}
        blurOnSubmit={false}
      />
    </View>
  );
});

const styles = StyleSheet.create({
  container: {
    flexDirection: 'row',
    marginHorizontal: spacing.lg,
    paddingVertical: spacing.xs,
  },
  border: {
    width: 3,
    backgroundColor: colors.green,
    borderRadius: 2,
    marginRight: spacing.md,
  },
  input: {
    flex: 1,
    fontFamily: fontFamily.document,
    fontSize: fontSize.md,
    lineHeight: fontSize.md * 1.5,
    color: colors.textPrimary,
    fontStyle: 'italic',
    minHeight: 24,
    padding: 0,
  },
});
