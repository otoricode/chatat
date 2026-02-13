// ParagraphBlock — basic text block
import React, { useRef, useEffect } from 'react';
import { TextInput, StyleSheet } from 'react-native';
import { colors, fontFamily, fontSize, spacing } from '@/theme';
import type { BlockProps } from '../types';

export const ParagraphBlock = React.memo(function ParagraphBlock({
  block,
  isActive,
  readOnly,
  onChange,
  onFocus,
  onSubmit,
  onBackspace,
  onSlashTrigger,
}: BlockProps) {
  const inputRef = useRef<TextInput>(null);

  useEffect(() => {
    if (isActive && inputRef.current) {
      inputRef.current.focus();
    }
  }, [isActive]);

  const handleChangeText = (text: string) => {
    // Detect slash command at start
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
      placeholder="Ketik sesuatu..."
      placeholderTextColor={colors.textMuted}
      blurOnSubmit={false}
    />
  );
});

const styles = StyleSheet.create({
  input: {
    fontFamily: fontFamily.document,
    fontSize: fontSize.md,
    lineHeight: fontSize.md * 1.5,
    color: colors.textPrimary,
    paddingVertical: spacing.xs,
    paddingHorizontal: spacing.lg,
    minHeight: 24,
  },
});
