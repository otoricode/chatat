// ChatInput — message input bar with send button
import React, { useState, useCallback } from 'react';
import {
  View,
  TextInput,
  Pressable,
  Text,
  StyleSheet,
} from 'react-native';
import { colors, fontSize, fontFamily, spacing } from '@/theme';

type Props = {
  onSend: (text: string) => void;
  replyTo?: { id: string; content: string } | null;
  onCancelReply?: () => void;
};

export function ChatInput({ onSend, replyTo, onCancelReply }: Props) {
  const [text, setText] = useState('');

  const handleSend = useCallback(() => {
    const trimmed = text.trim();
    if (!trimmed) return;
    onSend(trimmed);
    setText('');
  }, [text, onSend]);

  return (
    <View style={styles.wrapper}>
      {replyTo && (
        <View style={styles.replyBar}>
          <View style={styles.replyContent}>
            <View style={styles.replyIndicator} />
            <Text style={styles.replyText} numberOfLines={1}>
              {replyTo.content}
            </Text>
          </View>
          <Pressable onPress={onCancelReply} style={styles.cancelReply}>
            <Text style={styles.cancelIcon}>{'\u2715'}</Text>
          </Pressable>
        </View>
      )}
      <View style={styles.container}>
        <TextInput
          style={styles.input}
          value={text}
          onChangeText={setText}
          placeholder="Ketik pesan..."
          placeholderTextColor={colors.textMuted}
          multiline
          maxLength={4096}
        />
        {text.trim().length > 0 && (
          <Pressable
            style={({ pressed }) => [styles.sendButton, pressed && styles.sendPressed]}
            onPress={handleSend}
          >
            <Text style={styles.sendIcon}>{'\u27A4'}</Text>
          </Pressable>
        )}
      </View>
    </View>
  );
}

const styles = StyleSheet.create({
  wrapper: {
    backgroundColor: colors.background,
    borderTopWidth: 1,
    borderTopColor: colors.border,
  },
  replyBar: {
    flexDirection: 'row',
    alignItems: 'center',
    paddingHorizontal: spacing.lg,
    paddingVertical: spacing.sm,
    backgroundColor: colors.surface,
  },
  replyContent: {
    flex: 1,
    flexDirection: 'row',
    alignItems: 'center',
  },
  replyIndicator: {
    width: 3,
    height: 20,
    backgroundColor: colors.green,
    borderRadius: 2,
    marginRight: spacing.sm,
  },
  replyText: {
    fontFamily: fontFamily.ui,
    fontSize: fontSize.sm,
    color: colors.textMuted,
    flex: 1,
  },
  cancelReply: {
    padding: spacing.xs,
  },
  cancelIcon: {
    fontSize: 16,
    color: colors.textMuted,
  },
  container: {
    flexDirection: 'row',
    alignItems: 'flex-end',
    paddingHorizontal: spacing.md,
    paddingVertical: spacing.sm,
    gap: spacing.sm,
  },
  input: {
    flex: 1,
    backgroundColor: colors.inputBackground,
    borderRadius: 20,
    paddingHorizontal: spacing.lg,
    paddingVertical: spacing.sm,
    fontFamily: fontFamily.ui,
    fontSize: fontSize.md,
    color: colors.textPrimary,
    maxHeight: 120,
    borderWidth: 1,
    borderColor: colors.border,
  },
  sendButton: {
    width: 40,
    height: 40,
    borderRadius: 20,
    backgroundColor: colors.green,
    justifyContent: 'center',
    alignItems: 'center',
  },
  sendPressed: {
    opacity: 0.8,
  },
  sendIcon: {
    fontSize: 18,
    color: colors.background,
  },
});
