// Privacy Settings Screen — control who can see your info
import React, { useState, useEffect, useCallback } from 'react';
import { ScrollView, StyleSheet, Text, TouchableOpacity, View, ActivityIndicator } from 'react-native';
import { useTranslation } from 'react-i18next';
import { SettingSection, SettingRow } from '@/components/settings';
import { usersApi } from '@/services/api';
import type { PrivacySettings } from '@/services/api/users';
import { colors, fontFamily, fontSize, spacing } from '@/theme';

type Visibility = 'everyone' | 'contacts' | 'nobody';

function VisibilityPicker({
  label,
  value,
  onChange,
}: {
  label: string;
  value: Visibility;
  onChange: (v: Visibility) => void;
}) {
  const { t } = useTranslation();
  const options: { key: Visibility; label: string }[] = [
    { key: 'everyone', label: t('privacy.everyone') },
    { key: 'contacts', label: t('privacy.contacts') },
    { key: 'nobody', label: t('privacy.nobody') },
  ];

  return (
    <View style={pickerStyles.container}>
      <Text style={pickerStyles.label}>{label}</Text>
      <View style={pickerStyles.options}>
        {options.map((opt) => (
          <TouchableOpacity
            key={opt.key}
            style={[
              pickerStyles.option,
              value === opt.key && pickerStyles.optionActive,
            ]}
            onPress={() => onChange(opt.key)}
            activeOpacity={0.7}
          >
            <Text
              style={[
                pickerStyles.optionText,
                value === opt.key && pickerStyles.optionTextActive,
              ]}
            >
              {opt.label}
            </Text>
          </TouchableOpacity>
        ))}
      </View>
    </View>
  );
}

export function PrivacySettingsScreen() {
  const { t } = useTranslation();
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [settings, setSettings] = useState<PrivacySettings>({
    lastSeenVisibility: 'everyone',
    onlineVisibility: 'everyone',
    readReceipts: true,
    profilePhotoVisibility: 'everyone',
  });

  useEffect(() => {
    usersApi
      .getPrivacy()
      .then((res) => setSettings(res.data))
      .catch(() => {})
      .finally(() => setLoading(false));
  }, []);

  const update = useCallback(
    (patch: Partial<PrivacySettings>) => {
      const next = { ...settings, ...patch };
      setSettings(next);
      setSaving(true);
      usersApi
        .updatePrivacy(patch)
        .catch(() => setSettings(settings)) // revert on error
        .finally(() => setSaving(false));
    },
    [settings],
  );

  if (loading) {
    return (
      <View style={styles.center}>
        <ActivityIndicator color={colors.green} size="large" />
      </View>
    );
  }

  return (
    <ScrollView style={styles.container} contentContainerStyle={styles.content}>
      {saving && (
        <ActivityIndicator
          color={colors.green}
          size="small"
          style={styles.savingIndicator}
        />
      )}

      <SettingSection title={t('privacy.visibility')}>
        <VisibilityPicker
          label={t('privacy.lastSeen')}
          value={settings.lastSeenVisibility}
          onChange={(v) => update({ lastSeenVisibility: v })}
        />
        <VisibilityPicker
          label={t('privacy.onlineStatus')}
          value={settings.onlineVisibility}
          onChange={(v) => update({ onlineVisibility: v })}
        />
        <VisibilityPicker
          label={t('privacy.profilePhoto')}
          value={settings.profilePhotoVisibility}
          onChange={(v) => update({ profilePhotoVisibility: v })}
        />
      </SettingSection>

      <SettingSection title={t('privacy.messaging')}>
        <SettingRow
          label={t('privacy.readReceipts')}
          type="switch"
          switchValue={settings.readReceipts}
          onToggle={(v) => update({ readReceipts: v })}
          showDivider={false}
        />
        <Text style={styles.hint}>{t('privacy.readReceiptsHint')}</Text>
      </SettingSection>
    </ScrollView>
  );
}

const pickerStyles = StyleSheet.create({
  container: {
    paddingVertical: spacing.md,
    paddingHorizontal: spacing.lg,
    borderBottomWidth: StyleSheet.hairlineWidth,
    borderBottomColor: colors.border,
  },
  label: {
    fontFamily: fontFamily.ui,
    fontSize: fontSize.md,
    color: colors.textPrimary,
    marginBottom: spacing.sm,
  },
  options: {
    flexDirection: 'row',
    gap: spacing.sm,
  },
  option: {
    flex: 1,
    paddingVertical: spacing.sm,
    paddingHorizontal: spacing.md,
    borderRadius: 8,
    backgroundColor: colors.surface2,
    alignItems: 'center',
  },
  optionActive: {
    backgroundColor: colors.green,
  },
  optionText: {
    fontFamily: fontFamily.ui,
    fontSize: fontSize.sm,
    color: colors.textMuted,
  },
  optionTextActive: {
    color: colors.background,
    fontFamily: fontFamily.uiSemiBold,
  },
});

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: colors.background,
  },
  content: {
    padding: spacing.lg,
  },
  center: {
    flex: 1,
    justifyContent: 'center',
    alignItems: 'center',
    backgroundColor: colors.background,
  },
  savingIndicator: {
    position: 'absolute',
    top: spacing.md,
    right: spacing.lg,
  },
  hint: {
    fontFamily: fontFamily.ui,
    fontSize: fontSize.xs,
    color: colors.textMuted,
    paddingHorizontal: spacing.lg,
    paddingBottom: spacing.md,
  },
});
