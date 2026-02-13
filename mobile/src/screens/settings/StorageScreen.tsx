// Storage Screen — storage usage overview and cache management
import React, { useCallback, useState } from 'react';
import { View, Text, ScrollView, StyleSheet, TouchableOpacity, Alert } from 'react-native';
import { useTranslation } from 'react-i18next';
import AsyncStorage from '@react-native-async-storage/async-storage';
import { SettingSection, SettingRow } from '@/components/settings';
import { useSettingsStore } from '@/stores/settingsStore';
import { colors, fontFamily, fontSize, spacing } from '@/theme';

function formatBytes(bytes: number): string {
  if (bytes === 0) return '0 B';
  const k = 1024;
  const sizes = ['B', 'KB', 'MB', 'GB'];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return parseFloat((bytes / Math.pow(k, i)).toFixed(1)) + ' ' + sizes[i];
}

type StorageItem = {
  label: string;
  bytes: number;
  color: string;
};

export function StorageScreen() {
  const { t } = useTranslation();
  const autoDownload = useSettingsStore((s) => s.autoDownload);
  const updateAutoDownload = useSettingsStore((s) => s.updateAutoDownload);
  const [isClearing, setIsClearing] = useState(false);

  // Placeholder storage data — in production, calculate from actual storage
  const storageItems: StorageItem[] = [
    { label: t('settings.messages'), bytes: 0, color: colors.green },
    { label: t('settings.media'), bytes: 0, color: colors.blue },
    { label: t('settings.documents'), bytes: 0, color: colors.purple },
    { label: t('settings.cache'), bytes: 0, color: colors.yellow },
  ];

  const totalBytes = storageItems.reduce((sum, item) => sum + item.bytes, 0);

  const handleClearCache = useCallback(() => {
    Alert.alert(t('settings.clearCache'), t('settings.clearCacheConfirm'), [
      { text: t('common.cancel'), style: 'cancel' },
      {
        text: t('settings.clearCache'),
        style: 'destructive',
        onPress: async () => {
          setIsClearing(true);
          try {
            // Clear image cache and temp data
            // In production, clear actual image cache, temp files, etc.
            await AsyncStorage.multiRemove([
              'image-cache',
              'search-cache',
            ]).catch(() => {});
            Alert.alert(t('common.success'), t('settings.cacheCleared'));
          } catch {
            Alert.alert(t('common.error'), t('common.failed'));
          } finally {
            setIsClearing(false);
          }
        },
      },
    ]);
  }, [t]);

  return (
    <ScrollView style={styles.container} contentContainerStyle={styles.content}>
      {/* Storage Usage */}
      <SettingSection title={t('settings.storageUsage')}>
        <View style={styles.totalContainer}>
          <Text style={styles.totalLabel}>{t('settings.totalUsage')}</Text>
          <Text style={styles.totalValue}>{formatBytes(totalBytes)}</Text>
        </View>
        {storageItems.map((item, index) => (
          <View key={item.label}>
            <View style={styles.storageRow}>
              <View style={[styles.colorDot, { backgroundColor: item.color }]} />
              <Text style={styles.storageLabel}>{item.label}</Text>
              <Text style={styles.storageValue}>{formatBytes(item.bytes)}</Text>
            </View>
            {index < storageItems.length - 1 && <View style={styles.divider} />}
          </View>
        ))}
      </SettingSection>

      {/* Clear Cache */}
      <TouchableOpacity
        style={[styles.clearButton, isClearing && styles.clearButtonDisabled]}
        onPress={handleClearCache}
        disabled={isClearing}
        activeOpacity={0.7}
      >
        <Text style={styles.clearButtonText}>
          {isClearing ? t('common.loading') : t('settings.clearCache')}
        </Text>
      </TouchableOpacity>

      {/* Auto-download */}
      <SettingSection title={t('settings.autoDownload')}>
        <SettingRow
          label={t('settings.wifiAutoDownload')}
          type="switch"
          switchValue={autoDownload.wifiAll}
          onToggle={(v) => updateAutoDownload({ wifiAll: v })}
        />
        <SettingRow
          label={t('settings.cellularSmall')}
          type="switch"
          switchValue={autoDownload.cellularUnder5MB}
          onToggle={(v) => updateAutoDownload({ cellularUnder5MB: v })}
        />
        <SettingRow
          label={t('settings.cellularAsk')}
          type="switch"
          switchValue={autoDownload.cellularAsk}
          onToggle={(v) => updateAutoDownload({ cellularAsk: v })}
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
    paddingBottom: spacing.xxxl,
  },
  totalContainer: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    paddingHorizontal: spacing.lg,
    paddingVertical: spacing.md,
    borderBottomWidth: StyleSheet.hairlineWidth,
    borderBottomColor: colors.border,
  },
  totalLabel: {
    fontFamily: fontFamily.uiSemiBold,
    fontSize: fontSize.md,
    color: colors.textPrimary,
  },
  totalValue: {
    fontFamily: fontFamily.uiSemiBold,
    fontSize: fontSize.md,
    color: colors.green,
  },
  storageRow: {
    flexDirection: 'row',
    alignItems: 'center',
    paddingHorizontal: spacing.lg,
    paddingVertical: spacing.md,
  },
  colorDot: {
    width: 10,
    height: 10,
    borderRadius: 5,
    marginRight: spacing.md,
  },
  storageLabel: {
    flex: 1,
    fontFamily: fontFamily.ui,
    fontSize: fontSize.md,
    color: colors.textPrimary,
  },
  storageValue: {
    fontFamily: fontFamily.uiMedium,
    fontSize: fontSize.md,
    color: colors.textMuted,
  },
  divider: {
    height: StyleSheet.hairlineWidth,
    backgroundColor: colors.border,
    marginLeft: spacing.xxxl,
    marginRight: spacing.lg,
  },
  clearButton: {
    backgroundColor: 'transparent',
    borderRadius: 12,
    borderWidth: 1,
    borderColor: colors.red,
    paddingVertical: spacing.md,
    alignItems: 'center',
    marginBottom: spacing.lg,
  },
  clearButtonDisabled: {
    opacity: 0.5,
  },
  clearButtonText: {
    fontFamily: fontFamily.uiSemiBold,
    fontSize: fontSize.md,
    color: colors.red,
  },
});
