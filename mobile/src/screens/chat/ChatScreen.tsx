// Chat screen — individual chat view (placeholder)
import React from 'react';
import { View, Text, StyleSheet } from 'react-native';
import { SafeAreaView } from 'react-native-safe-area-context';
import type { NativeStackScreenProps } from '@react-navigation/native-stack';
import type { ChatStackParamList } from '@/navigation/types';
import { colors, fontSize, fontFamily } from '@/theme';

type Props = NativeStackScreenProps<ChatStackParamList, 'Chat'>;

export function ChatScreen({ route }: Props) {
  const { chatId, chatType } = route.params;

  return (
    <SafeAreaView style={styles.container}>
      <View style={styles.content}>
        <Text style={styles.text}>Chat: {chatId}</Text>
        <Text style={styles.subtext}>Type: {chatType}</Text>
      </View>
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
    justifyContent: 'center',
    alignItems: 'center',
  },
  text: {
    fontFamily: fontFamily.ui,
    fontSize: fontSize.lg,
    color: colors.textPrimary,
  },
  subtext: {
    fontFamily: fontFamily.ui,
    fontSize: fontSize.sm,
    color: colors.textMuted,
  },
});
