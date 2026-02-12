// Profile setup screen — name + emoji avatar picker
import React, { useState } from 'react';
import { View, Text, StyleSheet, TextInput, Pressable, ScrollView } from 'react-native';
import { SafeAreaView } from 'react-native-safe-area-context';
import type { NativeStackScreenProps } from '@react-navigation/native-stack';
import type { AuthStackParamList } from '@/navigation/types';
import { useAuthStore } from '@/stores/authStore';
import { colors, fontSize, spacing, fontFamily } from '@/theme';

type Props = NativeStackScreenProps<AuthStackParamList, 'ProfileSetup'>;

const AVATAR_EMOJIS = [
  '😀', '😎', '🤩', '🥳', '😇', '🤓', '🧑‍💻', '👨‍💼',
  '👩‍💼', '🧑‍🎨', '🧑‍🔬', '🧑‍🚀', '🦊', '🐱', '🐶', '🦁',
  '🐸', '🦉', '🐼', '🐨', '🌟', '🌈', '🔥', '💎',
  '🎯', '🎨', '🎸', '🎮', '📚', '🏆', '💡', '🚀',
];

export function ProfileSetupScreen(_props: Props) {
  const [name, setName] = useState('');
  const [avatar, setAvatar] = useState('😀');
  const [isLoading, setIsLoading] = useState(false);
  const login = useAuthStore((state) => state.login);

  const isValid = name.trim().length >= 2;

  const handleSubmit = async () => {
    if (!isValid) return;
    setIsLoading(true);
    try {
      // TODO: Call API to setup profile
      // For now, create a mock user and set auth
      login(
        { accessToken: 'mock-token', refreshToken: 'mock-refresh' },
        {
          id: 'temp-id',
          name: name.trim(),
          phone: '',
          avatar,
          status: '',
        },
        true,
      );
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <SafeAreaView style={styles.container}>
      <ScrollView contentContainerStyle={styles.content}>
        <Text style={styles.title}>Atur Profil</Text>
        <Text style={styles.description}>Pilih avatar dan masukkan namamu</Text>

        <View style={styles.avatarPreview}>
          <Text style={styles.avatarLarge}>{avatar}</Text>
        </View>

        <View style={styles.emojiGrid}>
          {AVATAR_EMOJIS.map((emoji) => (
            <Pressable
              key={emoji}
              style={[styles.emojiItem, avatar === emoji && styles.emojiSelected]}
              onPress={() => setAvatar(emoji)}
            >
              <Text style={styles.emojiText}>{emoji}</Text>
            </Pressable>
          ))}
        </View>

        <View style={styles.nameSection}>
          <Text style={styles.label}>Nama</Text>
          <TextInput
            style={styles.nameInput}
            value={name}
            onChangeText={setName}
            placeholder="Masukkan nama"
            placeholderTextColor={colors.textMuted}
            maxLength={50}
          />
        </View>

        <Pressable
          style={[styles.submitButton, !isValid && styles.buttonDisabled]}
          onPress={handleSubmit}
          disabled={!isValid || isLoading}
        >
          <Text style={styles.submitText}>
            {isLoading ? 'Menyimpan...' : 'Mulai'}
          </Text>
        </Pressable>
      </ScrollView>
    </SafeAreaView>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: colors.background,
  },
  content: {
    paddingHorizontal: spacing.xxl,
    paddingTop: spacing.xxxl,
    paddingBottom: spacing.xxxl,
  },
  title: {
    fontFamily: fontFamily.uiBold,
    fontSize: fontSize.xl,
    color: colors.textPrimary,
    textAlign: 'center',
    marginBottom: spacing.xs,
  },
  description: {
    fontFamily: fontFamily.ui,
    fontSize: fontSize.sm,
    color: colors.textMuted,
    textAlign: 'center',
    marginBottom: spacing.xxl,
  },
  avatarPreview: {
    alignSelf: 'center',
    width: 80,
    height: 80,
    borderRadius: 40,
    backgroundColor: colors.surface,
    justifyContent: 'center',
    alignItems: 'center',
    marginBottom: spacing.xxl,
  },
  avatarLarge: {
    fontSize: 48,
  },
  emojiGrid: {
    flexDirection: 'row',
    flexWrap: 'wrap',
    justifyContent: 'center',
    gap: spacing.sm,
    marginBottom: spacing.xxl,
  },
  emojiItem: {
    width: 44,
    height: 44,
    borderRadius: 22,
    justifyContent: 'center',
    alignItems: 'center',
    backgroundColor: colors.surface,
  },
  emojiSelected: {
    backgroundColor: colors.green,
  },
  emojiText: {
    fontSize: 24,
  },
  nameSection: {
    marginBottom: spacing.xxl,
  },
  label: {
    fontFamily: fontFamily.uiMedium,
    fontSize: fontSize.sm,
    color: colors.textPrimary,
    marginBottom: spacing.sm,
  },
  nameInput: {
    backgroundColor: colors.surface,
    borderRadius: 12,
    paddingHorizontal: spacing.lg,
    paddingVertical: spacing.md,
    fontFamily: fontFamily.ui,
    fontSize: fontSize.lg,
    color: colors.textPrimary,
    borderWidth: 1,
    borderColor: colors.border,
  },
  submitButton: {
    backgroundColor: colors.green,
    borderRadius: 12,
    paddingVertical: spacing.lg,
    alignItems: 'center',
  },
  buttonDisabled: {
    opacity: 0.4,
  },
  submitText: {
    fontFamily: fontFamily.uiSemiBold,
    fontSize: fontSize.md,
    color: colors.background,
  },
});
