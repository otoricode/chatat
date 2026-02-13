// Chat list screen — main chat tab
import React, { useCallback, useEffect, useState } from 'react';
import {
  View,
  FlatList,
  RefreshControl,
  StyleSheet,
  Alert,
} from 'react-native';
import { SafeAreaView } from 'react-native-safe-area-context';
import type { NativeStackScreenProps } from '@react-navigation/native-stack';
import type { ChatStackParamList } from '@/navigation/types';
import { Header } from '@/components/shared/Header';
import { FAB } from '@/components/shared/FAB';
import { EmptyState } from '@/components/shared/EmptyState';
import { ChatListItem } from '@/components/chat/ChatListItem';
import { useChatStore } from '@/stores/chatStore';
import { colors, spacing } from '@/theme';
import type { ChatListItem as ChatListItemType } from '@/types/chat';

type Props = NativeStackScreenProps<ChatStackParamList, 'ChatList'>;

export function ChatListScreen({ navigation }: Props) {
  const { chats, isLoading, fetchChats, pinChat, unpinChat, markAsRead } = useChatStore();
  const [refreshing, setRefreshing] = useState(false);

  useEffect(() => {
    fetchChats();
  }, [fetchChats]);

  const handleRefresh = useCallback(async () => {
    setRefreshing(true);
    await fetchChats();
    setRefreshing(false);
  }, [fetchChats]);

  const handleNewChat = () => {
    navigation.navigate('ContactList');
  };

  const handleSearchPress = () => {
    navigation.navigate('Search');
  };

  const handleChatPress = useCallback(
    (item: ChatListItemType) => {
      navigation.navigate('Chat', {
        chatId: item.chat.id,
        chatType: item.chat.type,
      });
    },
    [navigation],
  );

  const handleChatLongPress = useCallback(
    (item: ChatListItemType) => {
      const isPinned = item.chat.pinnedAt !== null;
      const options = [
        {
          text: isPinned ? 'Lepas Pin' : 'Pin Chat',
          onPress: () => (isPinned ? unpinChat(item.chat.id) : pinChat(item.chat.id)),
        },
        {
          text: 'Tandai Dibaca',
          onPress: () => markAsRead(item.chat.id),
        },
        { text: 'Batal', style: 'cancel' as const },
      ];

      Alert.alert('Opsi Chat', undefined, options);
    },
    [pinChat, unpinChat, markAsRead],
  );

  const renderItem = useCallback(
    ({ item }: { item: ChatListItemType }) => (
      <ChatListItem item={item} onPress={handleChatPress} onLongPress={handleChatLongPress} />
    ),
    [handleChatPress, handleChatLongPress],
  );

  const keyExtractor = useCallback((item: ChatListItemType) => item.chat.id, []);

  return (
    <SafeAreaView style={styles.container} edges={['top']}>
      <Header title="Chat" onSearchPress={handleSearchPress} />
      {chats.length === 0 && !isLoading ? (
        <View style={styles.content}>
          <EmptyState
            emoji="\u{1F4AC}"
            title="Belum ada chat"
            description="Mulai percakapan baru dengan kontak kamu"
          />
        </View>
      ) : (
        <FlatList
          data={chats}
          renderItem={renderItem}
          keyExtractor={keyExtractor}
          refreshControl={
            <RefreshControl
              refreshing={refreshing}
              onRefresh={handleRefresh}
              tintColor={colors.green}
              colors={[colors.green]}
            />
          }
          contentContainerStyle={chats.length === 0 ? styles.emptyList : undefined}
        />
      )}
      <FAB onPress={handleNewChat} />
    </SafeAreaView>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: colors.background,
  },
  content: {
    flex: 1,
    paddingHorizontal: spacing.lg,
  },
  emptyList: {
    flex: 1,
  },
});
