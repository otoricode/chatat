// HeadingBlock — H1, H2, H3
import React, { useRef, useEffect } from 'react';
import { TextInput, StyleSheet } from 'react-native';
import { colors, fontFamily, fontSize, spacing } from '@/theme';
import type { BlockProps } from '../types';

const headingConfig = {
  heading1: { size: fontSize.h1, family: fontFamily.documentBold, placeholder: 'Judul 1' },
  heading2: { size: fontSize.h2, family: fontFamily.documentBold, placeholder: 'Judul 2' },
  heading3: { size: fontSize.h3, family: fontFamily.documentMedium, placeholder: 'Judul 3' },
} as const;

export const HeadingBlock = React.memo(function HeadingBlock({
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
  const config = headingConfig[block.type as keyof typeof headingConfig] ?? headingConfig.heading1;

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
    <TextInput
      ref={inputRef}
      style={[styles.input, { fontSize: config.size, fontFamily: config.family }]}
      value={block.content}
      onChangeText={handleChangeText}
      onFocus={onFocus}
      onSubmitEditing={onSubmit}
      onKeyPress={handleKeyPress}
      multiline
      editable={!readOnly}
      placeholder={config.placeholder}
      placeholderTextColor={colors.textMuted}
      blurOnSubmit={false}
    />
  );
});

const styles = StyleSheet.create({
  input: {
    color: colors.textPrimary,
    paddingVertical: spacing.xs,
    paddingHorizontal: spacing.lg,
    lineHeight: 36,
    minHeight: 32,
  },
});
