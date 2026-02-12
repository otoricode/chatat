// Chat list screen — main chat tab
import React from 'react';
import { View, StyleSheet } from 'react-native';
import { SafeAreaView } from 'react-native-safe-area-context';
import type { NativeStackScreenProps } from '@react-navigation/native-stack';
import type { ChatStackParamList } from '@/navigation/types';
import { Header } from '@/components/shared/Header';
import { FAB } from '@/components/shared/FAB';
import { EmptyState } from '@/components/shared/EmptyState';
import { colors, spacing } from '@/theme';

type Props = NativeStackScreenProps<ChatStackParamList, 'ChatList'>;

export function ChatListScreen({ navigation }: Props) {
  const handleNewChat = () => {
    navigation.navigate('ContactList');
  };

  return (
    <SafeAreaView style={styles.container} edges={['top']}>
      <Header title="Chat" />
      <View style={styles.content}>
        <EmptyState
          emoji="💬"
          title="Belum ada chat"
          description="Mulai percakapan baru dengan kontak kamu"
        />
      </View>
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
});
