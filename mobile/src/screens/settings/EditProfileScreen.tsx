// EditProfile Screen — edit name, status, avatar
import React, { useCallback, useState } from 'react';
import {
  View,
  Text,
  StyleSheet,
  ScrollView,
  TextInput,
  TouchableOpacity,
  Alert,
  ActivityIndicator,
  KeyboardAvoidingView,
  Platform,
} from 'react-native';
import { useTranslation } from 'react-i18next';
import { useNavigation } from '@react-navigation/native';
import { useAuthStore } from '@/stores/authStore';
import { usersApi } from '@/services/api';
import { Avatar } from '@/components/ui/Avatar';
import { colors, fontFamily, fontSize, spacing } from '@/theme';

const MAX_NAME_LENGTH = 50;
const MAX_STATUS_LENGTH = 140;

const AVATAR_OPTIONS = [
  '😀', '😎', '🤓', '🧑', '👩', '👨', '🦊', '🐱',
  '🐶', '🦁', '🐯', '🐻', '🐼', '🐨', '🐸', '🦄',
  '🌟', '🌈', '🔥', '💎', '🎯', '🎨', '🎵', '🚀',
];

export function EditProfileScreen() {
  const { t } = useTranslation();
  const navigation = useNavigation();
  const user = useAuthStore((s) => s.user);
  const updateProfile = useAuthStore((s) => s.updateProfile);

  const [name, setName] = useState(user?.name ?? '');
  const [status, setStatus] = useState(user?.status ?? '');
  const [avatar, setAvatar] = useState(user?.avatar ?? '👤');
  const [showAvatarPicker, setShowAvatarPicker] = useState(false);
  const [isSaving, setIsSaving] = useState(false);

  const handleSave = useCallback(async () => {
    if (!name.trim()) {
      Alert.alert(t('common.error'), t('settings.nameRequired'));
      return;
    }

    setIsSaving(true);
    try {
      await usersApi.updateMe({
        name: name.trim(),
        status: status.trim(),
        avatar,
      });
      updateProfile({ name: name.trim(), status: status.trim(), avatar });
      navigation.goBack();
    } catch {
      Alert.alert(t('common.error'), t('common.saveFailed'));
    } finally {
      setIsSaving(false);
    }
  }, [name, status, avatar, updateProfile, navigation, t]);

  return (
    <KeyboardAvoidingView
      style={styles.container}
      behavior={Platform.OS === 'ios' ? 'padding' : undefined}
    >
      <ScrollView contentContainerStyle={styles.content}>
        {/* Avatar */}
        <TouchableOpacity
          style={styles.avatarContainer}
          onPress={() => setShowAvatarPicker(!showAvatarPicker)}
          activeOpacity={0.7}
        >
          <Avatar emoji={avatar} size="lg" />
          <View style={styles.cameraIcon}>
            <Text style={styles.cameraEmoji}>✏️</Text>
          </View>
        </TouchableOpacity>

        {/* Avatar Picker */}
        {showAvatarPicker && (
          <View style={styles.avatarPicker}>
            {AVATAR_OPTIONS.map((emoji) => (
              <TouchableOpacity
                key={emoji}
                style={[
                  styles.avatarOption,
                  avatar === emoji && styles.avatarOptionActive,
                ]}
                onPress={() => {
                  setAvatar(emoji);
                  setShowAvatarPicker(false);
                }}
              >
                <Text style={styles.avatarEmoji}>{emoji}</Text>
              </TouchableOpacity>
            ))}
          </View>
        )}

        {/* Name */}
        <View style={styles.inputSection}>
          <Text style={styles.inputLabel}>{t('auth.nameLabel')}</Text>
          <TextInput
            style={styles.input}
            value={name}
            onChangeText={(v) => setName(v.slice(0, MAX_NAME_LENGTH))}
            placeholder={t('auth.namePlaceholder')}
            placeholderTextColor={colors.textMuted}
            maxLength={MAX_NAME_LENGTH}
            autoCapitalize="words"
          />
          <Text style={styles.charCount}>
            {name.length}/{MAX_NAME_LENGTH}
          </Text>
        </View>

        {/* Status */}
        <View style={styles.inputSection}>
          <Text style={styles.inputLabel}>{t('settings.statusLabel')}</Text>
          <TextInput
            style={[styles.input, styles.multilineInput]}
            value={status}
            onChangeText={(v) => setStatus(v.slice(0, MAX_STATUS_LENGTH))}
            placeholder={t('settings.statusPlaceholder')}
            placeholderTextColor={colors.textMuted}
            maxLength={MAX_STATUS_LENGTH}
            multiline
          />
          <Text style={styles.charCount}>
            {status.length}/{MAX_STATUS_LENGTH}
          </Text>
        </View>

        {/* Phone (read-only) */}
        <View style={styles.inputSection}>
          <Text style={styles.inputLabel}>{t('auth.phoneLabel')}</Text>
          <View style={styles.readOnlyField}>
            <Text style={styles.readOnlyText}>{user?.phone ?? ''}</Text>
          </View>
          <Text style={styles.hint}>{t('settings.phoneCannotChange')}</Text>
        </View>

        {/* Save */}
        <TouchableOpacity
          style={[styles.saveButton, isSaving && styles.saveButtonDisabled]}
          onPress={handleSave}
          disabled={isSaving}
          activeOpacity={0.7}
        >
          {isSaving ? (
            <ActivityIndicator size="small" color={colors.background} />
          ) : (
            <Text style={styles.saveButtonText}>{t('common.save')}</Text>
          )}
        </TouchableOpacity>
      </ScrollView>
    </KeyboardAvoidingView>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: colors.background,
  },
  content: {
    padding: spacing.lg,
    paddingBottom: spacing.xxxl,
    alignItems: 'center',
  },
  avatarContainer: {
    marginBottom: spacing.xl,
    position: 'relative',
  },
  cameraIcon: {
    position: 'absolute',
    bottom: -2,
    right: -2,
    backgroundColor: colors.green,
    borderRadius: 12,
    width: 24,
    height: 24,
    justifyContent: 'center',
    alignItems: 'center',
  },
  cameraEmoji: {
    fontSize: 12,
  },
  avatarPicker: {
    flexDirection: 'row',
    flexWrap: 'wrap',
    justifyContent: 'center',
    backgroundColor: colors.surface,
    borderRadius: 12,
    padding: spacing.md,
    marginBottom: spacing.xl,
    gap: spacing.sm,
  },
  avatarOption: {
    width: 44,
    height: 44,
    borderRadius: 22,
    justifyContent: 'center',
    alignItems: 'center',
    backgroundColor: colors.surface2,
  },
  avatarOptionActive: {
    backgroundColor: colors.green,
  },
  avatarEmoji: {
    fontSize: 24,
  },
  inputSection: {
    width: '100%',
    marginBottom: spacing.lg,
  },
  inputLabel: {
    fontFamily: fontFamily.uiSemiBold,
    fontSize: fontSize.sm,
    color: colors.textMuted,
    textTransform: 'uppercase',
    letterSpacing: 0.5,
    marginBottom: spacing.sm,
  },
  input: {
    backgroundColor: colors.surface,
    borderRadius: 12,
    padding: spacing.lg,
    fontFamily: fontFamily.ui,
    fontSize: fontSize.md,
    color: colors.textPrimary,
  },
  multilineInput: {
    minHeight: 80,
    textAlignVertical: 'top',
  },
  charCount: {
    fontFamily: fontFamily.ui,
    fontSize: fontSize.xs,
    color: colors.textMuted,
    textAlign: 'right',
    marginTop: spacing.xs,
  },
  readOnlyField: {
    backgroundColor: colors.surface,
    borderRadius: 12,
    padding: spacing.lg,
    opacity: 0.6,
  },
  readOnlyText: {
    fontFamily: fontFamily.ui,
    fontSize: fontSize.md,
    color: colors.textPrimary,
  },
  hint: {
    fontFamily: fontFamily.ui,
    fontSize: fontSize.xs,
    color: colors.textMuted,
    marginTop: spacing.xs,
  },
  saveButton: {
    backgroundColor: colors.green,
    borderRadius: 12,
    paddingVertical: spacing.md,
    paddingHorizontal: spacing.xxxl,
    alignItems: 'center',
    marginTop: spacing.lg,
    width: '100%',
  },
  saveButtonDisabled: {
    opacity: 0.5,
  },
  saveButtonText: {
    fontFamily: fontFamily.uiSemiBold,
    fontSize: fontSize.md,
    color: colors.background,
  },
});
