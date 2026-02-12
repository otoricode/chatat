// Create group screen (placeholder)
import React from 'react';
import { View, Text, StyleSheet } from 'react-native';
import { SafeAreaView } from 'react-native-safe-area-context';
import type { NativeStackScreenProps } from '@react-navigation/native-stack';
import type { ChatStackParamList } from '@/navigation/types';
import { colors, fontSize, fontFamily } from '@/theme';

type Props = NativeStackScreenProps<ChatStackParamList, 'CreateGroup'>;

export function CreateGroupScreen(_props: Props) {
  return (
    <SafeAreaView style={styles.container}>
      <View style={styles.content}>
        <Text style={styles.text}>Buat Grup Baru</Text>
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
});
