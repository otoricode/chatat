// Chat screen — individual chat conversation view
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
import { Avatar } from '@/components/ui/Avatar';
import { MessageBubble } from '@/components/chat/MessageBubble';
import { ChatInput } from '@/components/chat/ChatInput';
import { DateSeparator } from '@/components/chat/DateSeparator';
import { useMessageStore } from '@/stores/messageStore';
import { useChatStore } from '@/stores/chatStore';
import { useAuthStore } from '@/stores/authStore';
import { chatsApi } from '@/services/api/chats';
import { isDifferentDay } from '@/lib/timeFormat';
import { formatLastSeen } from '@/lib/timeFormat';
import { colors, fontSize, fontFamily, spacing } from '@/theme';
import type { Message } from '@/types/chat';

type Props = NativeStackScreenProps<ChatStackParamList, 'Chat'>;

export function ChatScreen({ route, navigation }: Props) {
  const { chatId } = route.params;
  const currentUserId = useAuthStore((s) => s.user?.id);
  const { messages, isLoading, hasMore, fetchMessages, fetchMore, addMessage, deleteMessage } =
    useMessageStore();
  const { markAsRead } = useChatStore();

  const [replyTo, setReplyTo] = useState<{ id: string; content: string } | null>(null);
  const flatListRef = useRef<FlatList<Message>>(null);

  // Find other user info from chat store
  const chatItem = useChatStore((s) => s.chats.find((c) => c.chat.id === chatId));
  const otherUser = chatItem?.otherUser;

  const chatMessages = messages[chatId] ?? [];
  const chatHasMore = hasMore[chatId] ?? false;

  useEffect(() => {
    fetchMessages(chatId);
    markAsRead(chatId);
  }, [chatId, fetchMessages, markAsRead]);

  // Set header
  useEffect(() => {
    navigation.setOptions({
      headerShown: true,
      headerStyle: { backgroundColor: colors.headerBackground },
      headerTintColor: colors.textPrimary,
      headerTitle: () => (
        <Pressable
          style={headerStyles.titleContainer}
          onPress={() => navigation.navigate('ChatInfo', { chatId })}
        >
          <Avatar
            emoji={otherUser?.avatar || '\u{1F464}'}
            size="sm"
            online={chatItem?.isOnline}
          />
          <View>
            <Text style={headerStyles.name}>{otherUser?.name || 'Chat'}</Text>
            <Text style={headerStyles.status}>
              {otherUser
                ? formatLastSeen(otherUser.lastSeen, chatItem?.isOnline ?? false)
                : ''}
            </Text>
          </View>
        </Pressable>
      ),
    });
  }, [navigation, chatId, otherUser, chatItem]);

  const handleSend = useCallback(
    async (text: string) => {
      try {
        const res = await chatsApi.sendMessage(chatId, {
          content: text,
          replyToId: replyTo?.id ?? null,
          type: 'text',
        });
        addMessage(chatId, res.data.data);
        setReplyTo(null);
      } catch {
        Alert.alert('Gagal', 'Pesan gagal dikirim. Coba lagi.');
      }
    },
    [chatId, replyTo, addMessage],
  );

  const handleMessageLongPress = useCallback(
    (message: Message) => {
      const isSelf = message.senderId === currentUserId;
      const options: Array<{ text: string; onPress?: () => void; style?: 'cancel' | 'destructive' }> = [
        {
          text: 'Balas',
          onPress: () => setReplyTo({ id: message.id, content: message.content }),
        },
        {
          text: 'Salin',
          onPress: () => {
            // Clipboard would be imported in production
          },
        },
      ];

      if (isSelf) {
        options.push({
          text: 'Hapus untuk semua',
          style: 'destructive',
          onPress: async () => {
            try {
              await chatsApi.deleteMessage(chatId, message.id, true);
              deleteMessage(chatId, message.id);
            } catch {
              Alert.alert('Gagal', 'Pesan gagal dihapus.');
            }
          },
        });
      }

      options.push({
        text: 'Hapus untuk saya',
        style: 'destructive',
        onPress: async () => {
          try {
            await chatsApi.deleteMessage(chatId, message.id, false);
            deleteMessage(chatId, message.id);
          } catch {
            Alert.alert('Gagal', 'Pesan gagal dihapus.');
          }
        },
      });

      options.push({ text: 'Batal', style: 'cancel' });

      Alert.alert('Opsi Pesan', undefined, options);
    },
    [chatId, currentUserId, deleteMessage],
  );

  const handleEndReached = useCallback(() => {
    if (chatHasMore && !isLoading) {
      fetchMore(chatId);
    }
  }, [chatId, chatHasMore, isLoading, fetchMore]);

  const renderItem = useCallback(
    ({ item, index }: { item: Message; index: number }) => {
      const isSelf = item.senderId === currentUserId;
      const prevMessage = chatMessages[index + 1]; // inverted list; next index = older
      const showDate = !prevMessage || isDifferentDay(item.createdAt, prevMessage.createdAt);

      return (
        <>
          <MessageBubble
            message={item}
            isSelf={isSelf}
            onLongPress={handleMessageLongPress}
          />
          {showDate && <DateSeparator dateStr={item.createdAt} />}
        </>
      );
    },
    [currentUserId, chatMessages, handleMessageLongPress],
  );

  const keyExtractor = useCallback((item: Message) => item.id, []);

  return (
    <SafeAreaView style={styles.container} edges={['bottom']}>
      <KeyboardAvoidingView
        style={styles.flex}
        behavior={Platform.OS === 'ios' ? 'padding' : undefined}
        keyboardVerticalOffset={90}
      >
        <FlatList
          ref={flatListRef}
          data={chatMessages}
          renderItem={renderItem}
          keyExtractor={keyExtractor}
          inverted
          onEndReached={handleEndReached}
          onEndReachedThreshold={0.3}
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
