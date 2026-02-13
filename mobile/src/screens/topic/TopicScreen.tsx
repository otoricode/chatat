// Topic screen — topic conversation view with messages
import React, { useCallback, useEffect, useRef, useState } from 'react';
import {
  View,
  Text,
  FlatList,
  Pressable,
  Alert,
  KeyboardAvoidingView,
  Platform,
  StyleSheet,
} from 'react-native';
import { SafeAreaView } from 'react-native-safe-area-context';
import type { NativeStackScreenProps } from '@react-navigation/native-stack';
import type { ChatStackParamList } from '@/navigation/types';
import { MessageBubble } from '@/components/chat/MessageBubble';
import { ChatInput } from '@/components/chat/ChatInput';
import { DateSeparator } from '@/components/chat/DateSeparator';
import { useTopicStore } from '@/stores/topicStore';
import { useAuthStore } from '@/stores/authStore';
import { topicsApi } from '@/services/api/topics';
import { isDifferentDay } from '@/lib/timeFormat';
import { colors, fontSize, fontFamily, spacing } from '@/theme';
import type { TopicMessage, TopicDetail, Message } from '@/types/chat';

type Props = NativeStackScreenProps<ChatStackParamList, 'Topic'>;

// Adapt TopicMessage to Message shape for MessageBubble
function toMessage(tm: TopicMessage): Message {
  return {
    id: tm.id,
    chatId: tm.topicId,
    senderId: tm.senderId,
    content: tm.content,
    replyToId: tm.replyToId,
    type: tm.type,
    metadata: null,
    isDeleted: tm.isDeleted,
    deletedForAll: tm.deletedForAll,
    createdAt: tm.createdAt,
  };
}

export function TopicScreen({ route, navigation }: Props) {
  const { topicId } = route.params;
  const currentUserId = useAuthStore((s) => s.user?.id);
  const { messagesByTopic, fetchMessages, addMessage, deleteMessage } = useTopicStore();

  const [topicDetail, setTopicDetail] = useState<TopicDetail | null>(null);
  const [replyTo, setReplyTo] = useState<{ id: string; content: string } | null>(null);
  const [memberMap, setMemberMap] = useState<Record<string, { name: string; avatar: string }>>({});
  const flatListRef = useRef<FlatList<TopicMessage>>(null);

  const topicMessages = messagesByTopic[topicId] ?? [];

  // Fetch topic detail + messages
  useEffect(() => {
    fetchMessages(topicId);

    topicsApi.getById(topicId).then((res) => {
      const detail = res.data.data;
      setTopicDetail(detail);
      const map: Record<string, { name: string; avatar: string }> = {};
      for (const m of detail.members) {
        map[m.user.id] = { name: m.user.name, avatar: m.user.avatar };
      }
      setMemberMap(map);
    }).catch(() => {
      // silent
    });
  }, [topicId, fetchMessages]);

  // Set header
  useEffect(() => {
    const topic = topicDetail?.topic;
    if (!topic) return;

    const memberCount = topicDetail?.members.length ?? 0;

    navigation.setOptions({
      headerShown: true,
      headerStyle: { backgroundColor: colors.headerBackground },
      headerTintColor: colors.textPrimary,
      headerTitle: () => (
        <Pressable
          style={headerStyles.titleContainer}
          onPress={() => navigation.navigate('TopicInfo', { topicId })}
        >
          <View style={headerStyles.iconWrap}>
            <Text style={headerStyles.icon}>{topic.icon || '📌'}</Text>
          </View>
          <View>
            <Text style={headerStyles.name} numberOfLines={1}>
              {topic.name}
            </Text>
            <Text style={headerStyles.status}>
              {memberCount} anggota
            </Text>
          </View>
        </Pressable>
      ),
    });
  }, [navigation, topicDetail, topicId]);

  const handleSend = useCallback(
    async (text: string) => {
      try {
        const res = await topicsApi.sendMessage(topicId, {
          content: text,
          replyToId: replyTo?.id ?? undefined,
          type: 'text',
        });
        addMessage(topicId, res.data.data);
        setReplyTo(null);
      } catch {
        Alert.alert('Gagal', 'Pesan gagal dikirim. Coba lagi.');
      }
    },
    [topicId, replyTo, addMessage],
  );

  const handleMessageLongPress = useCallback(
    (message: Message) => {
      const isSelf = message.senderId === currentUserId;
      const options: Array<{
        text: string;
        onPress?: () => void;
        style?: 'cancel' | 'destructive';
      }> = [
        {
          text: 'Balas',
          onPress: () => setReplyTo({ id: message.id, content: message.content }),
        },
      ];

      if (isSelf) {
        options.push({
          text: 'Hapus',
          style: 'destructive',
          onPress: async () => {
            try {
              await topicsApi.deleteMessage(topicId, message.id, true);
              deleteMessage(topicId, message.id);
            } catch {
              Alert.alert('Gagal', 'Pesan gagal dihapus.');
            }
          },
        });
      }

      options.push({ text: 'Batal', style: 'cancel' });
      Alert.alert('Opsi Pesan', undefined, options);
    },
    [topicId, currentUserId, deleteMessage],
  );

  const renderItem = useCallback(
    ({ item, index }: { item: TopicMessage; index: number }) => {
      const msg = toMessage(item);
      const isSelf = item.senderId === currentUserId;
      const prevMessage = topicMessages[index + 1]; // inverted
      const showDate = !prevMessage || isDifferentDay(item.createdAt, prevMessage.createdAt);

      const senderName = !isSelf
        ? memberMap[item.senderId]?.name ?? undefined
        : undefined;
      const showSenderName = !isSelf && (
        !prevMessage || prevMessage.senderId !== item.senderId
      );

      return (
        <>
          <MessageBubble
            message={msg}
            isSelf={isSelf}
            onLongPress={handleMessageLongPress}
            senderName={showSenderName ? senderName : undefined}
          />
          {showDate && <DateSeparator dateStr={item.createdAt} />}
        </>
      );
    },
    [currentUserId, topicMessages, handleMessageLongPress, memberMap],
  );

  const keyExtractor = useCallback((item: TopicMessage) => item.id, []);

  return (
    <SafeAreaView style={styles.container} edges={['bottom']}>
      <KeyboardAvoidingView
        style={styles.flex}
        behavior={Platform.OS === 'ios' ? 'padding' : undefined}
        keyboardVerticalOffset={90}
      >
        <FlatList
          ref={flatListRef}
          data={topicMessages}
          renderItem={renderItem}
          keyExtractor={keyExtractor}
          inverted
          contentContainerStyle={styles.messageList}
        />
        <ChatInput
          onSend={handleSend}
          replyTo={replyTo}
          onCancelReply={() => setReplyTo(null)}
        />
      </KeyboardAvoidingView>
    </SafeAreaView>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: colors.background,
  },
  flex: {
    flex: 1,
  },
  messageList: {
    paddingVertical: spacing.sm,
  },
});

const headerStyles = StyleSheet.create({
  titleContainer: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: spacing.sm,
  },
  iconWrap: {
    width: 32,
    height: 32,
    borderRadius: spacing.sm,
    backgroundColor: colors.surface2,
    justifyContent: 'center',
    alignItems: 'center',
  },
  icon: {
    fontSize: fontSize.md,
  },
  name: {
    fontFamily: fontFamily.uiSemiBold,
    fontSize: fontSize.md,
    color: colors.textPrimary,
  },
  status: {
    fontFamily: fontFamily.ui,
    fontSize: fontSize.xs,
    color: colors.textMuted,
  },
});
