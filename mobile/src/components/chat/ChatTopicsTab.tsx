// Topics tab content — shows topics belonging to a group chat
import React, { useCallback, useEffect } from 'react';
import {
  View,
  Text,
  FlatList,
  TouchableOpacity,
  ActivityIndicator,
  StyleSheet,
} from 'react-native';
import { useNavigation } from '@react-navigation/native';
import type { NativeStackNavigationProp } from '@react-navigation/native-stack';
import type { ChatStackParamList } from '@/navigation/types';
import { EmptyState } from '@/components/shared/EmptyState';
import { useTopicStore } from '@/stores/topicStore';
import { useChatStore } from '@/stores/chatStore';
import { useTranslation } from 'react-i18next';
import { formatMessageTime } from '@/lib/timeFormat';
import { colors, fontSize, fontFamily, spacing } from '@/theme';
import type { TopicListItem } from '@/types/chat';

type NavigationProp = NativeStackNavigationProp<ChatStackParamList>;

interface ChatTopicsTabProps {
  chatId: string;
}

export function ChatTopicsTab({ chatId }: ChatTopicsTabProps) {
  const { t } = useTranslation();
  const navigation = useNavigation<NavigationProp>();
  const { topicsByChat, isLoading, fetchTopics } = useTopicStore();
  const chatItem = useChatStore((s) => s.chats.find((c) => c.chat.id === chatId));
  const chatType = chatItem?.chat.type ?? 'group';

  const topics = topicsByChat[chatId] ?? [];

  useEffect(() => {
    fetchTopics(chatId);
  }, [chatId, fetchTopics]);

  const handleTopicPress = useCallback(
    (topicId: string) => {
      navigation.navigate('Topic', { topicId });
    },
    [navigation],
  );

  const handleCreate = useCallback(() => {
    navigation.navigate('CreateTopic', { chatId, chatType });
  }, [navigation, chatId, chatType]);

  const renderItem = useCallback(
    ({ item }: { item: TopicListItem }) => (
      <TouchableOpacity
        style={styles.topicRow}
        onPress={() => handleTopicPress(item.topic.id)}
      >
        <View style={styles.topicIcon}>
          <Text style={styles.iconText}>{item.topic.icon || '\u{1F4CC}'}</Text>
        </View>
        <View style={styles.topicInfo}>
          <View style={styles.topicHeader}>
            <Text style={styles.topicName} numberOfLines={1}>
              {item.topic.name}
            </Text>
            {item.lastMessage && (
              <Text style={styles.time}>
                {formatMessageTime(item.lastMessage.createdAt)}
              </Text>
            )}
          </View>
          <View style={styles.topicSub}>
            <Text style={styles.lastMessage} numberOfLines={1}>
              {item.lastMessage?.isDeleted
                ? t('chat.messageDeleted')
                : item.lastMessage?.content ?? t('chat.noChats')}
            </Text>
            <Text style={styles.memberCount}>
              {t('group.memberCount', { count: item.memberCount })}
            </Text>
          </View>
        </View>
      </TouchableOpacity>
    ),
    [handleTopicPress, t],
  );

  const keyExtractor = useCallback((item: TopicListItem) => item.topic.id, []);

  if (isLoading && topics.length === 0) {
    return (
      <View style={styles.center}>
        <ActivityIndicator color={colors.green} />
      </View>
    );
  }

  return (
    <View style={styles.container}>
      {topics.length === 0 ? (
        <View style={styles.emptyContainer}>
          <EmptyState
            emoji="\u{1F4CC}"
            title={t('topic.noTopics')}
            description={t('topic.noTopicsDesc')}
          />
          <TouchableOpacity style={styles.createButton} onPress={handleCreate}>
            <Text style={styles.createButtonText}>{t('topic.newTopic')}</Text>
          </TouchableOpacity>
        </View>
      ) : (
        <>
          <FlatList
            data={topics}
            renderItem={renderItem}
            keyExtractor={keyExtractor}
            contentContainerStyle={styles.listContent}
          />
          <TouchableOpacity style={styles.fab} onPress={handleCreate}>
            <Text style={styles.fabIcon}>+</Text>
          </TouchableOpacity>
        </>
      )}
    </View>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: colors.background,
  },
  center: {
    flex: 1,
    justifyContent: 'center',
    alignItems: 'center',
  },
  emptyContainer: {
    flex: 1,
    justifyContent: 'center',
    alignItems: 'center',
    padding: spacing.xl,
  },
  listContent: {
    paddingVertical: spacing.sm,
  },
  topicRow: {
    flexDirection: 'row',
    alignItems: 'center',
    paddingHorizontal: spacing.md,
    paddingVertical: spacing.sm,
    gap: spacing.md,
  },
  topicIcon: {
    width: 48,
    height: 48,
    borderRadius: 24,
    backgroundColor: colors.surface2,
    justifyContent: 'center',
    alignItems: 'center',
  },
  iconText: {
    fontSize: 22,
  },
  topicInfo: {
    flex: 1,
  },
  topicHeader: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
  },
  topicName: {
    fontFamily: fontFamily.uiSemiBold,
    fontSize: fontSize.md,
    color: colors.textPrimary,
    flex: 1,
  },
  time: {
    fontFamily: fontFamily.ui,
    fontSize: fontSize.xs,
    color: colors.textMuted,
    marginLeft: spacing.sm,
  },
  topicSub: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    marginTop: 2,
  },
  lastMessage: {
    fontFamily: fontFamily.ui,
    fontSize: fontSize.sm,
    color: colors.textMuted,
    flex: 1,
  },
  memberCount: {
    fontFamily: fontFamily.ui,
    fontSize: fontSize.xs,
    color: colors.textMuted,
    marginLeft: spacing.sm,
  },
  createButton: {
    marginTop: spacing.lg,
    backgroundColor: colors.green,
    paddingHorizontal: spacing.xl,
    paddingVertical: spacing.md,
    borderRadius: 12,
  },
  createButtonText: {
    fontFamily: fontFamily.uiSemiBold,
    fontSize: fontSize.md,
    color: colors.white,
  },
  fab: {
    position: 'absolute',
    bottom: spacing.lg,
    right: spacing.lg,
    width: 56,
    height: 56,
    borderRadius: 28,
    backgroundColor: colors.green,
    justifyContent: 'center',
    alignItems: 'center',
    elevation: 4,
    shadowColor: '#000',
    shadowOffset: { width: 0, height: 2 },
    shadowOpacity: 0.25,
    shadowRadius: 4,
  },
  fabIcon: {
    fontSize: 28,
    color: colors.white,
    lineHeight: 30,
  },
});
