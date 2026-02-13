// Chat screen — individual chat conversation view with tabs
import React, { useCallback, useEffect, useMemo, useRef, useState } from 'react';
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
import { AttachmentPicker } from '@/components/chat/AttachmentPicker';
import type { PickedMedia } from '@/components/chat/AttachmentPicker';
import { DateSeparator } from '@/components/chat/DateSeparator';
import { ChatTabBar } from '@/components/chat/ChatTabBar';
import type { ChatTab } from '@/components/chat/ChatTabBar';
import { ChatDocumentsTab } from '@/components/chat/ChatDocumentsTab';
import { ChatTopicsTab } from '@/components/chat/ChatTopicsTab';
import { useMessageStore } from '@/stores/messageStore';
import { useChatStore } from '@/stores/chatStore';
import { useAuthStore } from '@/stores/authStore';
import { chatsApi } from '@/services/api/chats';
import { mediaApi } from '@/services/api/media';
import { isDifferentDay } from '@/lib/timeFormat';
import { formatLastSeen } from '@/lib/timeFormat';
import { useTypingIndicator } from '@/hooks/useTypingIndicator';
import { useTranslation } from 'react-i18next';
import { wsClient } from '@/services/ws';
import { colors, fontSize, fontFamily, spacing } from '@/theme';
import type { Message, MediaResponse } from '@/types/chat';

type Props = NativeStackScreenProps<ChatStackParamList, 'Chat'>;

export function ChatScreen({ route, navigation }: Props) {
  const { chatId, chatType } = route.params;
  const { t } = useTranslation();
  const currentUserId = useAuthStore((s) => s.user?.id);
  const { messages, isLoading, hasMore, fetchMessages, fetchMore, addMessage, deleteMessage } =
    useMessageStore();
  const { markAsRead } = useChatStore();

  const [activeTab, setActiveTab] = useState<ChatTab>('chat');
  const [replyTo, setReplyTo] = useState<{ id: string; content: string } | null>(null);
  const [showAttachment, setShowAttachment] = useState(false);
  const [, setUploadProgress] = useState<Record<string, number>>({});
  const flatListRef = useRef<FlatList<Message>>(null);

  // Find other user info from chat store
  const chatItem = useChatStore((s) => s.chats.find((c) => c.chat.id === chatId));
  const otherUser = chatItem?.otherUser;
  const isGroup = chatType === 'group';

  // For groups: member map (userId -> User) for sender names
  const [memberMap, setMemberMap] = useState<Record<string, { name: string; avatar: string }>>({});


  // Tab definitions: personal has 2 tabs, group has 3 tabs
  const tabs = useMemo(() => {
    const base = [
      { key: 'chat' as ChatTab, label: t('chat.tabChat'), icon: '\u{1F4AC}' },
      { key: 'documents' as ChatTab, label: t('chat.tabDocuments'), icon: '\u{1F4C4}' },
    ];
    if (isGroup) {
      base.push({ key: 'topics' as ChatTab, label: t('chat.tabTopics'), icon: '\u{1F4CC}' });
    }
    return base;
  }, [isGroup, t]);
  const chatMessages = messages[chatId] ?? [];
  const chatHasMore = hasMore[chatId] ?? false;

  // Typing indicator
  const { typingText, sendTyping } = useTypingIndicator(chatId);

  useEffect(() => {
    fetchMessages(chatId);
    markAsRead(chatId);

    // Send read receipt via WebSocket
    const lastMsg = (messages[chatId] ?? [])[0];
    if (lastMsg) {
      wsClient.send('read_receipt', {
        chatId,
        lastReadMessageId: lastMsg.id,
      });
    }
  }, [chatId, fetchMessages, markAsRead, messages]);

  // Load group member info
  useEffect(() => {
    if (isGroup) {
      chatsApi.getGroupInfo(chatId).then((res) => {
        const map: Record<string, { name: string; avatar: string }> = {};
        for (const m of res.data.data.members) {
          map[m.user.id] = { name: m.user.name, avatar: m.user.avatar };
        }
        setMemberMap(map);
      }).catch(() => {
        // Silent fail for member info
      });
    }
  }, [chatId, isGroup]);

  // Set header
  useEffect(() => {
    if (isGroup) {
      const chat = chatItem?.chat;
      const memberCount = Object.keys(memberMap).length;
      navigation.setOptions({
        headerShown: true,
        headerStyle: { backgroundColor: colors.headerBackground },
        headerTintColor: colors.textPrimary,
        headerTitle: () => (
          <Pressable
            style={headerStyles.titleContainer}
            onPress={() => navigation.navigate('ChatInfo', { chatId, chatType: 'group' })}
          >
            <Avatar emoji={chat?.icon || '\u{1F465}'} size="sm" />
            <View>
              <Text style={headerStyles.name}>{chat?.name || t('group.groupInfo')}</Text>
              <Text style={[headerStyles.status, typingText ? headerStyles.typingStatus : undefined]}>
                {typingText || (memberCount > 0 ? t('group.memberCount', { count: memberCount }) : '')}
              </Text>
            </View>
          </Pressable>
        ),
      });
    } else {
      navigation.setOptions({
        headerShown: true,
        headerStyle: { backgroundColor: colors.headerBackground },
        headerTintColor: colors.textPrimary,
        headerTitle: () => (
          <Pressable
            style={headerStyles.titleContainer}
            onPress={() => navigation.navigate('ChatInfo', { chatId, chatType: 'personal' })}
          >
            <Avatar
              emoji={otherUser?.avatar || '\u{1F464}'}
              size="sm"
              online={chatItem?.isOnline}
            />
            <View>
              <Text style={headerStyles.name}>{otherUser?.name || t('chat.title')}</Text>
              <Text style={[headerStyles.status, typingText ? headerStyles.typingStatus : undefined]}>
                {typingText || (otherUser
                  ? formatLastSeen(otherUser.lastSeen, chatItem?.isOnline ?? false, t)
                  : '')}
              </Text>
            </View>
          </Pressable>
        ),
      });
    }
  }, [navigation, chatId, otherUser, chatItem, isGroup, memberMap, typingText]);

  const handleSend = useCallback(
    async (text: string) => {
      // Stop typing indicator on send
      sendTyping(false);
      try {
        const res = await chatsApi.sendMessage(chatId, {
          content: text,
          replyToId: replyTo?.id ?? null,
          type: 'text',
        });
        addMessage(chatId, res.data.data);
        setReplyTo(null);
      } catch {
        Alert.alert(t('common.failed'), t('chat.sendFailed'));
      }
    },
    [chatId, replyTo, addMessage, sendTyping],
  );

  const handleAttach = useCallback(
    async (picked: PickedMedia) => {
      const tempId = `temp_${Date.now()}`;
      const messageType = picked.type === 'image' ? 'image' : 'file';

      try {
        // Upload media
        const uploadRes = await mediaApi.upload(
          picked.uri,
          picked.filename,
          picked.mimeType,
          {
            contextType: 'chat',
            contextId: chatId,
            onProgress: (progress) => {
              setUploadProgress((prev) => ({ ...prev, [tempId]: progress }));
            },
          },
        );

        const media = uploadRes.data.data;

        // Send message with media metadata
        const res = await chatsApi.sendMessage(chatId, {
          content: media.filename,
          replyToId: null,
          type: messageType,
          metadata: {
            id: media.id,
            type: media.type,
            filename: media.filename,
            contentType: media.contentType,
            size: media.size,
            width: media.width,
            height: media.height,
            url: media.url,
            thumbnailURL: media.thumbnailURL,
          },
        });

        addMessage(chatId, res.data.data);
      } catch {
        Alert.alert(t('common.failed'), t('media.sendFailed'));
      } finally {
        setUploadProgress((prev) => {
          const next = { ...prev };
          delete next[tempId];
          return next;
        });
      }
    },
    [chatId, addMessage],
  );

  const handleImagePress = useCallback(
    (media: MediaResponse) => {
      navigation.navigate('ImageViewer', {
        url: media.url,
        filename: media.filename,
      });
    },
    [navigation],
  );

  const handleMessageLongPress = useCallback(
    (message: Message) => {
      const isSelf = message.senderId === currentUserId;
      const options: Array<{ text: string; onPress?: () => void; style?: 'cancel' | 'destructive' }> = [
        {
          text: t('chat.reply'),
          onPress: () => setReplyTo({ id: message.id, content: message.content }),
        },
        {
          text: t('common.copy'),
          onPress: () => {
            // Clipboard would be imported in production
          },
        },
      ];

      if (isSelf) {
        options.push({
          text: t('chat.deleteForAll'),
          style: 'destructive',
          onPress: async () => {
            try {
              await chatsApi.deleteMessage(chatId, message.id, true);
              deleteMessage(chatId, message.id);
            } catch {
              Alert.alert(t('common.failed'), t('chat.deleteFailed'));
            }
          },
        });
      }

      options.push({
        text: t('chat.deleteForMe'),
        style: 'destructive',
        onPress: async () => {
          try {
            await chatsApi.deleteMessage(chatId, message.id, false);
            deleteMessage(chatId, message.id);
          } catch {
            Alert.alert(t('common.failed'), t('chat.deleteFailed'));
          }
        },
      });

      options.push({ text: t('common.cancel'), style: 'cancel' });

      Alert.alert(t('chat.messageOptions'), undefined, options);
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

      // For group chats: show sender name if different from previous
      const senderName = isGroup && !isSelf
        ? memberMap[item.senderId]?.name ?? undefined
        : undefined;
      const showSenderName = isGroup && !isSelf && (
        !prevMessage || prevMessage.senderId !== item.senderId
      );

      return (
        <>
          <MessageBubble
            message={item}
            isSelf={isSelf}
            onLongPress={handleMessageLongPress}
            senderName={showSenderName ? senderName : undefined}
            onImagePress={handleImagePress}
          />
          {showDate && <DateSeparator dateStr={item.createdAt} />}
        </>
      );
    },
    [currentUserId, chatMessages, handleMessageLongPress, handleImagePress, isGroup, memberMap],
  );

  const keyExtractor = useCallback((item: Message) => item.id, []);

  return (
    <SafeAreaView style={styles.container} edges={['bottom']}>
      <ChatTabBar tabs={tabs} activeTab={activeTab} onTabChange={setActiveTab} />
      {activeTab === 'chat' && (
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
            removeClippedSubviews
            maxToRenderPerBatch={15}
            windowSize={10}
            initialNumToRender={20}
            onEndReached={handleEndReached}
            onEndReachedThreshold={0.3}
            contentContainerStyle={styles.messageList}
          />
          <ChatInput
            onSend={handleSend}
            onTyping={() => sendTyping(true)}
            onAttach={() => setShowAttachment(true)}
            replyTo={replyTo}
            onCancelReply={() => setReplyTo(null)}
          />
          <AttachmentPicker
            visible={showAttachment}
            onClose={() => setShowAttachment(false)}
            onPick={handleAttach}
          />
        </KeyboardAvoidingView>
      )}
      {activeTab === 'documents' && (
        <ChatDocumentsTab chatId={chatId} />
      )}
      {activeTab === 'topics' && isGroup && (
        <ChatTopicsTab chatId={chatId} />
      )}
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
  typingStatus: {
    color: colors.green,
    fontStyle: 'italic',
  },
});
