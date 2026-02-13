// ChecklistBlock — checkbox + text
import React, { useRef, useEffect } from 'react';
import { View, TextInput, Pressable, StyleSheet } from 'react-native';
import { useTranslation } from 'react-i18next';
import { colors, fontFamily, fontSize, spacing } from '@/theme';
import type { BlockProps } from '../types';

export const ChecklistBlock = React.memo(function ChecklistBlock({
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

  const handleToggle = () => {
    if (readOnly) return;
    onChange({ checked: !block.checked });
  };

  return (
    <View style={styles.row}>
      <Pressable onPress={handleToggle} style={styles.checkboxWrapper}>
        <View style={[styles.checkbox, block.checked && styles.checked]}>
          {block.checked && <View style={styles.checkmark} />}
        </View>
      </Pressable>
      <TextInput
        ref={inputRef}
        style={[styles.input, block.checked && styles.strikethrough]}
        value={block.content}
        onChangeText={handleChangeText}
        onFocus={onFocus}
        onSubmitEditing={onSubmit}
        onKeyPress={handleKeyPress}
        multiline
        editable={!readOnly}
        placeholder={t('editor.todo')}
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
  checkboxWrapper: {
    padding: 2,
    marginRight: spacing.sm,
    marginTop: 2,
  },
  checkbox: {
    width: 18,
    height: 18,
    borderRadius: 4,
    borderWidth: 2,
    borderColor: colors.textMuted,
    justifyContent: 'center',
    alignItems: 'center',
  },
  checked: {
    backgroundColor: colors.green,
    borderColor: colors.green,
  },
  checkmark: {
    width: 8,
    height: 4,
    borderLeftWidth: 2,
    borderBottomWidth: 2,
    borderColor: colors.background,
    transform: [{ rotate: '-45deg' }],
    marginTop: -1,
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
  strikethrough: {
    textDecorationLine: 'line-through',
    color: colors.textMuted,
  },
});
