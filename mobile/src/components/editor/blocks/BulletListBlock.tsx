// BulletListBlock — unordered list item
import React, { useRef, useEffect } from 'react';
import { View, TextInput, Text, StyleSheet } from 'react-native';
import { colors, fontFamily, fontSize, spacing } from '@/theme';
import type { BlockProps } from '../types';

export const BulletListBlock = React.memo(function BulletListBlock({
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
    <View style={styles.row}>
      <Text style={styles.bullet}>{'\u2022'}</Text>
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
        placeholder="Item daftar"
        placeholderTextColor={colors.textMuted}
        blurOnSubmit={false}
      />
    </View>
  );
});

const styles = StyleSheet.create({
  row: {
    flexDirection: 'row',
    alignItems: 'flex-start',
    paddingHorizontal: spacing.lg,
    paddingVertical: spacing.xs,
  },
  bullet: {
    fontFamily: fontFamily.document,
    fontSize: fontSize.md,
    color: colors.textPrimary,
    marginRight: spacing.sm,
    marginTop: 2,
    width: 16,
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
