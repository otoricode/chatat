// Notification Settings Screen — toggle notification preferences
import React from 'react';
import { ScrollView, StyleSheet } from 'react-native';
import { useTranslation } from 'react-i18next';
import { useSettingsStore } from '@/stores/settingsStore';
import { SettingSection, SettingRow } from '@/components/settings';
import { colors, spacing } from '@/theme';

export function NotificationSettingsScreen() {
  const { t } = useTranslation();
  const notifications = useSettingsStore((s) => s.notifications);
  const updateNotifications = useSettingsStore((s) => s.updateNotifications);

  return (
    <ScrollView style={styles.container} contentContainerStyle={styles.content}>
      <SettingSection title={t('settings.messageNotifications')}>
        <SettingRow
          label={t('settings.showPreview')}
          type="switch"
          switchValue={notifications.showPreview}
          onToggle={(v) => updateNotifications({ showPreview: v })}
        />
        <SettingRow
          label={t('settings.sound')}
          type="switch"
          switchValue={notifications.soundEnabled}
          onToggle={(v) => updateNotifications({ soundEnabled: v })}
        />
        <SettingRow
          label={t('settings.vibration')}
          type="switch"
          switchValue={notifications.vibrationEnabled}
          onToggle={(v) => updateNotifications({ vibrationEnabled: v })}
          showDivider={false}
        />
      </SettingSection>

      <SettingSection title={t('settings.groupNotifications')}>
        <SettingRow
          label={t('settings.groupAlerts')}
          type="switch"
          switchValue={notifications.groupAlerts}
          onToggle={(v) => updateNotifications({ groupAlerts: v })}
          showDivider={false}
        />
      </SettingSection>
    </ScrollView>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: colors.background,
  },
  content: {
    padding: spacing.lg,
  },
});
