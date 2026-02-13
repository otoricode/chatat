// Topic list screen — shows topics for a chat
import React, { useCallback, useEffect } from 'react';
import {
  View,
  Text,
  FlatList,
  TouchableOpacity,
  ActivityIndicator,
  StyleSheet,
} from 'react-native';
import { SafeAreaView } from 'react-native-safe-area-context';
import type { NativeStackScreenProps } from '@react-navigation/native-stack';
import type { ChatStackParamList } from '@/navigation/types';
import { EmptyState } from '@/components/shared/EmptyState';
import { useTopicStore } from '@/stores/topicStore';
import { useChatStore } from '@/stores/chatStore';
import { formatMessageTime } from '@/lib/timeFormat';
import { colors, fontSize, fontFamily, spacing } from '@/theme';
import type { TopicListItem } from '@/types/chat';

type Props = NativeStackScreenProps<ChatStackParamList, 'TopicList'>;

export function TopicListScreen({ route, navigation }: Props) {
  const { chatId } = route.params;
  const { topicsByChat, isLoading, fetchTopics } = useTopicStore();
  const chatItem = useChatStore((s) => s.chats.find((c) => c.chat.id === chatId));
  const chatType = chatItem?.chat.type ?? 'personal';

  const topics = topicsByChat[chatId] ?? [];

  useEffect(() => {
    fetchTopics(chatId);
  }, [chatId, fetchTopics]);

  // Set up header with create button
  useEffect(() => {
    navigation.setOptions({
      headerRight: () => (
        <TouchableOpacity
          onPress={() => navigation.navigate('CreateTopic', { chatId, chatType })}
          style={headerStyles.addButton}
        >
          <Text style={headerStyles.addIcon}>+</Text>
        </TouchableOpacity>
      ),
    });
  }, [navigation, chatId, chatType]);

  const renderItem = useCallback(
    ({ item }: { item: TopicListItem }) => (
      <TouchableOpacity
        style={styles.topicRow}
        onPress={() => navigation.navigate('Topic', { topicId: item.topic.id })}
      >
        <View style={styles.topicIcon}>
          <Text style={styles.iconText}>{item.topic.icon || '📌'}</Text>
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
                ? 'Pesan dihapus'
                : item.lastMessage?.content ?? 'Belum ada pesan'}
            </Text>
            <Text style={styles.memberCount}>
              {item.memberCount} anggota
            </Text>
          </View>
        </View>
      </TouchableOpacity>
    ),
    [navigation],
  );

  const keyExtractor = useCallback((item: TopicListItem) => item.topic.id, []);

  if (isLoading && topics.length === 0) {
    return (
      <SafeAreaView style={styles.container}>
        <View style={styles.center}>
          <ActivityIndicator color={colors.green} />
        </View>
      </SafeAreaView>
    );
  }

  return (
    <SafeAreaView style={styles.container} edges={['bottom']}>
      {topics.length === 0 ? (
        <View style={styles.emptyContainer}>
          <EmptyState
            emoji="📌"
            title="Belum ada topik"
            description="Buat topik untuk diskusi terfokus dalam chat ini"
          />
          <TouchableOpacity
            style={styles.createButton}
            onPress={() => navigation.navigate('CreateTopic', { chatId, chatType })}
          >
            <Text style={styles.createButtonText}>Buat Topik</Text>
          </TouchableOpacity>
        </View>
      ) : (
        <FlatList
          data={topics}
          renderItem={renderItem}
          keyExtractor={keyExtractor}
          contentContainerStyle={styles.listContent}
        />
      )}
    </SafeAreaView>
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
  },
  createButton: {
    backgroundColor: colors.green,
    paddingHorizontal: spacing.xl,
    paddingVertical: spacing.md,
    borderRadius: spacing.sm,
    marginTop: spacing.lg,
  },
  createButtonText: {
    fontFamily: fontFamily.uiSemiBold,
    fontSize: fontSize.md,
    color: colors.white,
  },
  listContent: {
    paddingVertical: spacing.sm,
  },
  topicRow: {
    flexDirection: 'row',
    alignItems: 'center',
    paddingHorizontal: spacing.lg,
    paddingVertical: spacing.md,
    borderBottomWidth: 1,
    borderBottomColor: colors.border,
  },
  topicIcon: {
    width: 44,
    height: 44,
    borderRadius: spacing.sm,
    backgroundColor: colors.surface2,
    justifyContent: 'center',
    alignItems: 'center',
  },
  iconText: {
    fontSize: fontSize.xl,
  },
  topicInfo: {
    flex: 1,
    marginLeft: spacing.md,
  },
  topicHeader: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    marginBottom: 2,
  },
  topicName: {
    fontFamily: fontFamily.uiSemiBold,
    fontSize: fontSize.md,
    color: colors.textPrimary,
    flex: 1,
    marginRight: spacing.sm,
  },
  time: {
    fontFamily: fontFamily.ui,
    fontSize: fontSize.xs,
    color: colors.textMuted,
  },
  topicSub: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
  },
  lastMessage: {
    fontFamily: fontFamily.ui,
    fontSize: fontSize.sm,
    color: colors.textMuted,
    flex: 1,
    marginRight: spacing.sm,
  },
  memberCount: {
    fontFamily: fontFamily.ui,
    fontSize: fontSize.xs,
    color: colors.textMuted,
  },
});

const headerStyles = StyleSheet.create({
  addButton: {
    width: 32,
    height: 32,
    borderRadius: 16,
    backgroundColor: colors.green,
    justifyContent: 'center',
    alignItems: 'center',
  },
  addIcon: {
    fontSize: fontSize.xl,
    color: colors.white,
    fontWeight: '600',
    marginTop: -1,
  },
});
