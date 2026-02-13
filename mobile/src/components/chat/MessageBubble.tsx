// MessageBubble — chat message bubble component
import React from 'react';
import { View, Text, Pressable, StyleSheet } from 'react-native';
import { colors, fontSize, fontFamily, spacing } from '@/theme';
import { formatMessageTime } from '@/lib/timeFormat';
import type { Message, DeliveryStatus } from '@/types/chat';

type Props = {
  message: Message;
  isSelf: boolean;
  onLongPress?: (message: Message) => void;
  onSwipeReply?: (message: Message) => void;
};

function getStatusIcon(status?: DeliveryStatus): string {
  switch (status) {
    case 'read':
      return '\u2713\u2713'; // ✓✓
    case 'delivered':
      return '\u2713\u2713';
    case 'sent':
    default:
      return '\u2713';
  }
}

export function MessageBubble({ message, isSelf, onLongPress }: Props) {
  if (message.isDeleted) {
    return (
      <View style={[styles.container, isSelf ? styles.selfContainer : styles.otherContainer]}>
        <View
          style={[
            styles.bubble,
            isSelf ? styles.selfBubble : styles.otherBubble,
            styles.deletedBubble,
          ]}
        >
          <Text style={styles.deletedText}>{'\u{1F6AB}'} Pesan dihapus</Text>
        </View>
      </View>
    );
  }

  const isForwarded = message.metadata ? Boolean((message.metadata as Record<string, unknown>).forwarded) : false;

  return (
    <View style={[styles.container, isSelf ? styles.selfContainer : styles.otherContainer]}>
      <Pressable
        style={({ pressed }) => [
          styles.bubble,
          isSelf ? styles.selfBubble : styles.otherBubble,
          pressed && styles.pressed,
        ]}
        onLongPress={() => onLongPress?.(message)}
      >
        {isForwarded && (
          <Text style={styles.forwardedLabel}>{'\u{27A1}'} Diteruskan</Text>
        )}

        {message.replyToId && (
          <View style={styles.replyPreview}>
            <View style={styles.replyBar} />
            <Text style={styles.replyText} numberOfLines={1}>
              Balasan pesan
            </Text>
          </View>
        )}

        <Text style={isSelf ? styles.selfText : styles.otherText}>{message.content}</Text>

        <View style={styles.meta}>
          <Text style={styles.time}>{formatMessageTime(message.createdAt)}</Text>
          {isSelf && <Text style={styles.status}>{getStatusIcon()}</Text>}
        </View>
      </Pressable>
    </View>
  );
}

const styles = StyleSheet.create({
  container: {
    paddingHorizontal: spacing.sm,
    paddingVertical: 2,
  },
  selfContainer: {
    alignItems: 'flex-end',
  },
  otherContainer: {
    alignItems: 'flex-start',
  },
  bubble: {
    maxWidth: '80%',
    paddingHorizontal: spacing.md,
    paddingTop: spacing.sm,
    paddingBottom: spacing.xs,
    borderRadius: 12,
  },
  selfBubble: {
    backgroundColor: colors.bubbleSelf,
    borderBottomRightRadius: 4,
  },
  otherBubble: {
    backgroundColor: colors.bubbleOther,
    borderBottomLeftRadius: 4,
  },
  pressed: {
    opacity: 0.8,
  },
  deletedBubble: {
    opacity: 0.6,
  },
  deletedText: {
    fontFamily: fontFamily.ui,
    fontSize: fontSize.sm,
    color: colors.textMuted,
    fontStyle: 'italic',
  },
  forwardedLabel: {
    fontFamily: fontFamily.ui,
    fontSize: fontSize.xs,
    color: colors.textMuted,
    fontStyle: 'italic',
    marginBottom: 2,
  },
  replyPreview: {
    flexDirection: 'row',
    alignItems: 'center',
    backgroundColor: 'rgba(255, 255, 255, 0.05)',
    borderRadius: 6,
    padding: spacing.xs,
    marginBottom: spacing.xs,
  },
  replyBar: {
    width: 3,
    height: '100%',
    backgroundColor: colors.green,
    borderRadius: 2,
    marginRight: spacing.xs,
    minHeight: 16,
  },
  replyText: {
    fontFamily: fontFamily.ui,
    fontSize: fontSize.xs,
    color: colors.textMuted,
    flex: 1,
  },
  selfText: {
    fontFamily: fontFamily.ui,
    fontSize: fontSize.md,
    color: colors.bubbleSelfText,
  },
  otherText: {
    fontFamily: fontFamily.ui,
    fontSize: fontSize.md,
    color: colors.bubbleOtherText,
  },
  meta: {
    flexDirection: 'row',
    justifyContent: 'flex-end',
    alignItems: 'center',
    gap: 4,
    marginTop: 2,
  },
  time: {
    fontFamily: fontFamily.ui,
    fontSize: 10,
    color: colors.textMuted,
  },
  status: {
    fontSize: 10,
    color: colors.textMuted,
  },
});
