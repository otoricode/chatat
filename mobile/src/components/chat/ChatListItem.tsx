// ChatListItem — single chat item in the chat list
import React, { useCallback } from 'react';
import { View, Text, Pressable, StyleSheet } from 'react-native';
import { useTranslation } from 'react-i18next';
import { Avatar } from '@/components/ui/Avatar';
import { Badge } from '@/components/ui/Badge';
import { colors, fontSize, fontFamily, spacing } from '@/theme';
import { formatChatListTime } from '@/lib/timeFormat';
import type { ChatListItem as ChatListItemType } from '@/types/chat';

type Props = {
  item: ChatListItemType;
  onPress: (item: ChatListItemType) => void;
  onLongPress: (item: ChatListItemType) => void;
};

export function ChatListItem({ item, onPress, onLongPress }: Props) {
  const { t } = useTranslation();
  const { chat, lastMessage, unreadCount, otherUser, isOnline } = item;

  const handlePress = useCallback(() => onPress(item), [item, onPress]);
  const handleLongPress = useCallback(() => onLongPress(item), [item, onLongPress]);

  // Display name: other user's name for personal chats, chat name for groups
  const displayName = chat.type === 'personal' && otherUser ? otherUser.name : chat.name;

  // Avatar emoji
  const avatarEmoji =
    chat.type === 'personal' && otherUser ? otherUser.avatar || '\u{1F464}' : chat.icon || '\u{1F465}';

  // Last message preview
  const preview = lastMessage
    ? lastMessage.isDeleted
      ? `\u{1F6AB} ${t('chat.messageDeleted')}`
      : lastMessage.content.length > 60
        ? lastMessage.content.slice(0, 60) + '...'
        : lastMessage.content
    : t('chat.noMessages');

  // Time
  const timeText = lastMessage ? formatChatListTime(lastMessage.createdAt, t) : '';

  const isPinned = chat.pinnedAt !== null;

  return (
    <Pressable
      style={({ pressed }) => [styles.container, pressed && styles.pressed]}
      onPress={handlePress}
      onLongPress={handleLongPress}
    >
      <Avatar emoji={avatarEmoji} size="md" online={chat.type === 'personal' ? isOnline : undefined} />

      <View style={styles.content}>
        <View style={styles.topRow}>
          <Text style={styles.name} numberOfLines={1}>
            {displayName || 'Chat'}
          </Text>
          <View style={styles.timeRow}>
            {isPinned && <Text style={styles.pinIcon}>{'\u{1F4CC}'}</Text>}
            <Text style={[styles.time, unreadCount > 0 && styles.timeUnread]}>{timeText}</Text>
          </View>
        </View>
        <View style={styles.bottomRow}>
          <Text style={styles.preview} numberOfLines={1}>
            {preview}
          </Text>
          {unreadCount > 0 && <Badge count={unreadCount} variant="unread" />}
        </View>
      </View>
    </Pressable>
  );
}

const styles = StyleSheet.create({
  container: {
    flexDirection: 'row',
    alignItems: 'center',
    paddingHorizontal: spacing.lg,
    paddingVertical: spacing.md,
    gap: spacing.md,
  },
  pressed: {
    backgroundColor: colors.surface,
  },
  content: {
    flex: 1,
  },
  topRow: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    marginBottom: 2,
  },
  name: {
    flex: 1,
    fontFamily: fontFamily.uiSemiBold,
    fontSize: fontSize.md,
    color: colors.textPrimary,
    marginRight: spacing.sm,
  },
  timeRow: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: 4,
  },
  pinIcon: {
    fontSize: 10,
  },
  time: {
    fontFamily: fontFamily.ui,
    fontSize: fontSize.xs,
    color: colors.textMuted,
  },
  timeUnread: {
    color: colors.green,
  },
  bottomRow: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
  },
  preview: {
    flex: 1,
    fontFamily: fontFamily.ui,
    fontSize: fontSize.sm,
    color: colors.textMuted,
    marginRight: spacing.sm,
  },
});
