// Settings Screen — main settings page
import React, { useCallback } from 'react';
import { ScrollView, View, Text, StyleSheet, TouchableOpacity, Alert } from 'react-native';
import { useTranslation } from 'react-i18next';
import { useNavigation } from '@react-navigation/native';
import type { NativeStackNavigationProp } from '@react-navigation/native-stack';
import { SafeAreaView } from 'react-native-safe-area-context';
import { useAuthStore } from '@/stores/authStore';
import { notificationsApi } from '@/services/api';
import { getCurrentLanguage } from '@/i18n';
import { SettingSection, SettingRow } from '@/components/settings';
import { Avatar } from '@/components/ui/Avatar';
import type { ChatStackParamList } from '@/navigation/types';
import { colors, fontFamily, fontSize, spacing } from '@/theme';

type Nav = NativeStackNavigationProp<ChatStackParamList>;

function getLanguageLabel(code: string): string {
  const labels: Record<string, string> = {
    id: 'Bahasa Indonesia',
    en: 'English',
    ar: '\u0627\u0644\u0639\u0631\u0628\u064a\u0629',
  };
  return labels[code] ?? code;
}

export function SettingsScreen() {
  const { t } = useTranslation();
  const navigation = useNavigation<Nav>();
  const user = useAuthStore((s) => s.user);
  const logout = useAuthStore((s) => s.logout);

  const handleLogout = useCallback(() => {
    Alert.alert(t('settings.logoutConfirm'), t('settings.logoutMessage'), [
      { text: t('common.cancel'), style: 'cancel' },
      {
        text: t('settings.logout'),
        style: 'destructive',
        onPress: async () => {
          try {
            // Unregister push token (best effort)
            const token = useAuthStore.getState().accessToken;
            if (token) {
              await notificationsApi.unregisterDevice(token).catch(() => {});
            }
          } catch {
            // ignore
          }
          // Clear auth state — navigates via auth state change
          logout();
        },
      },
    ]);
  }, [t, logout]);

  return (
    <SafeAreaView style={styles.container} edges={['top']}>
      <ScrollView contentContainerStyle={styles.content}>
        {/* Profile Header */}
        <TouchableOpacity
          style={styles.profileHeader}
          onPress={() => navigation.navigate('EditProfile')}
          activeOpacity={0.7}
        >
          <Avatar emoji={user?.avatar ?? '👤'} size="lg" />
          <View style={styles.profileInfo}>
            <Text style={styles.profileName} numberOfLines={1}>
              {user?.name ?? ''}
            </Text>
            <Text style={styles.profilePhone}>{user?.phone ?? ''}</Text>
            <Text style={styles.profileStatus} numberOfLines={1}>
              {user?.status || t('settings.noStatus')}
            </Text>
          </View>
          <Text style={styles.chevron}>›</Text>
        </TouchableOpacity>

        {/* Account Section */}
        <SettingSection title={t('settings.account')}>
          <SettingRow
            icon="🌐"
            label={t('settings.language')}
            value={getLanguageLabel(getCurrentLanguage())}
            onPress={() => navigation.navigate('Language')}
          />
          <SettingRow
            icon="🔔"
            label={t('settings.notifications')}
            onPress={() => navigation.navigate('NotificationSettings')}
          />
          <SettingRow
            icon="🔒"
            label={t('privacy.title')}
            onPress={() => navigation.navigate('PrivacySettings')}
          />
          <SettingRow
            icon="☁️"
            label={t('backup.title')}
            onPress={() => navigation.navigate('Backup')}
            showDivider={false}
          />
        </SettingSection>

        {/* Storage Section */}
        <SettingSection title={t('settings.storage')}>
          <SettingRow
            icon="💾"
            label={t('settings.storageUsage')}
            onPress={() => navigation.navigate('Storage')}
            showDivider={false}
          />
        </SettingSection>

        {/* About Section */}
        <SettingSection title={t('settings.about')}>
          <SettingRow
            icon="ℹ️"
            label={t('settings.about')}
            value="v1.0.0"
            onPress={() => navigation.navigate('About')}
            showDivider={false}
          />
        </SettingSection>

        {/* Logout */}
        <TouchableOpacity
          style={styles.logoutButton}
          onPress={handleLogout}
          activeOpacity={0.7}
        >
          <Text style={styles.logoutIcon}>🚪</Text>
          <Text style={styles.logoutText}>{t('settings.logout')}</Text>
        </TouchableOpacity>

        <Text style={styles.version}>Chatat v1.0.0</Text>
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
    padding: spacing.lg,
    paddingBottom: spacing.xxxl,
  },
  profileHeader: {
    flexDirection: 'row',
    alignItems: 'center',
    backgroundColor: colors.surface,
    borderRadius: 12,
    padding: spacing.lg,
    marginBottom: spacing.xl,
  },
  profileInfo: {
    flex: 1,
    marginLeft: spacing.md,
  },
  profileName: {
    fontFamily: fontFamily.uiSemiBold,
    fontSize: fontSize.lg,
    color: colors.textPrimary,
  },
  profilePhone: {
    fontFamily: fontFamily.ui,
    fontSize: fontSize.sm,
    color: colors.textMuted,
    marginTop: 2,
  },
  profileStatus: {
    fontFamily: fontFamily.ui,
    fontSize: fontSize.sm,
    color: colors.textMuted,
    marginTop: 2,
  },
  chevron: {
    fontFamily: fontFamily.ui,
    fontSize: fontSize.h1,
    color: colors.textMuted,
    marginLeft: spacing.sm,
  },
  logoutButton: {
    flexDirection: 'row',
    alignItems: 'center',
    backgroundColor: colors.surface,
    borderRadius: 12,
    padding: spacing.lg,
    marginBottom: spacing.lg,
  },
  logoutIcon: {
    fontSize: 20,
    marginRight: spacing.md,
  },
  logoutText: {
    fontFamily: fontFamily.uiSemiBold,
    fontSize: fontSize.md,
    color: colors.red,
  },
  version: {
    fontFamily: fontFamily.ui,
    fontSize: fontSize.sm,
    color: colors.textMuted,
    textAlign: 'center',
  },
});
