// NumberedListBlock — ordered list item
import React, { useRef, useEffect } from 'react';
import { View, TextInput, Text, StyleSheet } from 'react-native';
import { useTranslation } from 'react-i18next';
import { colors, fontFamily, fontSize, spacing } from '@/theme';
import type { BlockProps } from '../types';

interface NumberedListBlockProps extends BlockProps {
  index: number;
}

export const NumberedListBlock = React.memo(function NumberedListBlock({
  block,
  isActive,
  readOnly,
  onChange,
  onFocus,
  onSubmit,
  onBackspace,
  onSlashTrigger,
  index,
}: NumberedListBlockProps) {
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
    <View style={styles.row}>
      <Text style={styles.number}>{index}.</Text>
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
        placeholder={t('editor.listItem')}
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
  number: {
    fontFamily: fontFamily.document,
    fontSize: fontSize.md,
    color: colors.textMuted,
    marginRight: spacing.sm,
    marginTop: 2,
    minWidth: 20,
    textAlign: 'right',
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
