// Backup Screen — Cloud backup management
// Accessible from ChatList settings or profile
import React, { useCallback, useEffect, useState } from 'react';
import {
  View,
  Text,
  StyleSheet,
  ScrollView,
  TouchableOpacity,
  ActivityIndicator,
  Alert,
  Platform,
} from 'react-native';
import { useTranslation } from 'react-i18next';
import { useBackupStore } from '@/stores/backupStore';
import { colors, fontFamily, fontSize, spacing } from '@/theme';

function formatBytes(bytes: number): string {
  if (bytes === 0) return '0 B';
  const k = 1024;
  const sizes = ['B', 'KB', 'MB', 'GB'];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return parseFloat((bytes / Math.pow(k, i)).toFixed(1)) + ' ' + sizes[i];
}

function formatDate(dateStr: string): string {
  const d = new Date(dateStr);
  return d.toLocaleDateString(undefined, {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  });
}

export function BackupScreen() {
  const { t } = useTranslation();
  const {
    latestBackup,
    history,
    isLoading,
    isBackingUp,
    isRestoring,
    progress,
    error,
    platform,
    fetchLatest,
    fetchHistory,
    startBackup,
    startRestore,
    clearError,
  } = useBackupStore();

  const [showHistory, setShowHistory] = useState(false);

  useEffect(() => {
    fetchLatest();
    fetchHistory();
  }, [fetchLatest, fetchHistory]);

  useEffect(() => {
    if (error) {
      Alert.alert(t('common.error'), error, [
        { text: t('common.ok'), onPress: clearError },
      ]);
    }
  }, [error, clearError, t]);

  const handleBackup = useCallback(() => {
    Alert.alert(t('backup.confirmBackupTitle'), t('backup.confirmBackupMessage'), [
      { text: t('common.cancel'), style: 'cancel' },
      { text: t('backup.backupNow'), onPress: startBackup },
    ]);
  }, [startBackup, t]);

  const handleRestore = useCallback(() => {
    Alert.alert(t('backup.confirmRestoreTitle'), t('backup.confirmRestoreMessage'), [
      { text: t('common.cancel'), style: 'cancel' },
      {
        text: t('backup.restore'),
        style: 'destructive',
        onPress: startRestore,
      },
    ]);
  }, [startRestore, t]);

  const isProcessing = isBackingUp || isRestoring;
  const platformLabel =
    platform === 'icloud' ? 'iCloud' : 'Google Drive';
  const platformIcon = Platform.OS === 'ios' ? '☁️' : '📁';

  return (
    <ScrollView style={styles.container} contentContainerStyle={styles.content}>
      {/* Platform Info */}
      <View style={styles.platformCard}>
        <Text style={styles.platformIcon}>{platformIcon}</Text>
        <View style={styles.platformInfo}>
          <Text style={styles.platformLabel}>{platformLabel}</Text>
          <Text style={styles.platformHint}>
            {t('backup.platformHint', { platform: platformLabel })}
          </Text>
        </View>
      </View>

      {/* Last Backup Info */}
      <View style={styles.section}>
        <Text style={styles.sectionTitle}>{t('backup.lastBackup')}</Text>
        <View style={styles.card}>
          {latestBackup ? (
            <>
              <View style={styles.infoRow}>
                <Text style={styles.infoLabel}>{t('backup.date')}</Text>
                <Text style={styles.infoValue}>
                  {formatDate(latestBackup.createdAt)}
                </Text>
              </View>
              <View style={styles.divider} />
              <View style={styles.infoRow}>
                <Text style={styles.infoLabel}>{t('backup.size')}</Text>
                <Text style={styles.infoValue}>
                  {formatBytes(latestBackup.sizeBytes)}
                </Text>
              </View>
              <View style={styles.divider} />
              <View style={styles.infoRow}>
                <Text style={styles.infoLabel}>{t('backup.status')}</Text>
                <Text
                  style={[
                    styles.infoValue,
                    latestBackup.status === 'completed'
                      ? styles.statusSuccess
                      : styles.statusFailed,
                  ]}
                >
                  {latestBackup.status === 'completed'
                    ? t('backup.statusCompleted')
                    : t('backup.statusFailed')}
                </Text>
              </View>
            </>
          ) : (
            <Text style={styles.emptyText}>{t('backup.noBackupYet')}</Text>
          )}
        </View>
      </View>

      {/* Progress */}
      {isProcessing && progress && (
        <View style={styles.progressCard}>
          <ActivityIndicator size="small" color={colors.green} />
          <View style={styles.progressInfo}>
            <Text style={styles.progressLabel}>
              {isBackingUp ? t('backup.backingUp') : t('backup.restoring')}
            </Text>
            <View style={styles.progressBarBg}>
              <View
                style={[
                  styles.progressBarFill,
                  { width: `${Math.round(progress.progress * 100)}%` },
                ]}
              />
            </View>
            <Text style={styles.progressPercent}>
              {Math.round(progress.progress * 100)}%
            </Text>
          </View>
        </View>
      )}

      {/* Actions */}
      <View style={styles.section}>
        <TouchableOpacity
          style={[styles.primaryButton, isProcessing && styles.buttonDisabled]}
          onPress={handleBackup}
          disabled={isProcessing}
          activeOpacity={0.7}
        >
          {isBackingUp ? (
            <ActivityIndicator size="small" color={colors.background} />
          ) : (
            <Text style={styles.primaryButtonText}>
              {t('backup.backupNow')}
            </Text>
          )}
        </TouchableOpacity>

        <TouchableOpacity
          style={[styles.secondaryButton, isProcessing && styles.buttonDisabled]}
          onPress={handleRestore}
          disabled={isProcessing}
          activeOpacity={0.7}
        >
          {isRestoring ? (
            <ActivityIndicator size="small" color={colors.green} />
          ) : (
            <Text style={styles.secondaryButtonText}>
              {t('backup.restoreFromBackup')}
            </Text>
          )}
        </TouchableOpacity>
      </View>

      {/* Backup History */}
      <View style={styles.section}>
        <TouchableOpacity
          style={styles.historyHeader}
          onPress={() => setShowHistory(!showHistory)}
          activeOpacity={0.7}
        >
          <Text style={styles.sectionTitle}>{t('backup.history')}</Text>
          <Text style={styles.chevron}>{showHistory ? '▲' : '▼'}</Text>
        </TouchableOpacity>

        {showHistory && (
          <View style={styles.card}>
            {isLoading ? (
              <ActivityIndicator
                size="small"
                color={colors.green}
                style={styles.loader}
              />
            ) : history.length > 0 ? (
              history.map((record, index) => (
                <View key={record.id}>
                  {index > 0 && <View style={styles.divider} />}
                  <View style={styles.historyItem}>
                    <View style={styles.historyLeft}>
                      <Text style={styles.historyDate}>
                        {formatDate(record.createdAt)}
                      </Text>
                      <Text style={styles.historyMeta}>
                        {record.platform === 'google_drive'
                          ? 'Google Drive'
                          : 'iCloud'}{' '}
                        · {formatBytes(record.sizeBytes)}
                      </Text>
                    </View>
                    <Text
                      style={[
                        styles.historyStatus,
                        record.status === 'completed'
                          ? styles.statusSuccess
                          : styles.statusFailed,
                      ]}
                    >
                      {record.status === 'completed' ? '✓' : '✗'}
                    </Text>
                  </View>
                </View>
              ))
            ) : (
              <Text style={styles.emptyText}>{t('backup.noHistory')}</Text>
            )}
          </View>
        )}
      </View>
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
  platformCard: {
    flexDirection: 'row',
    alignItems: 'center',
    backgroundColor: colors.surface,
    borderRadius: 12,
    padding: spacing.lg,
    marginBottom: spacing.lg,
  },
  platformIcon: {
    fontSize: 32,
    marginRight: spacing.md,
  },
  platformInfo: {
    flex: 1,
  },
  platformLabel: {
    fontFamily: fontFamily.uiSemiBold,
    fontSize: fontSize.md,
    color: colors.textPrimary,
  },
  platformHint: {
    fontFamily: fontFamily.ui,
    fontSize: fontSize.sm,
    color: colors.textMuted,
    marginTop: 2,
  },
  section: {
    marginBottom: spacing.lg,
  },
  sectionTitle: {
    fontFamily: fontFamily.uiSemiBold,
    fontSize: fontSize.sm,
    color: colors.textMuted,
    textTransform: 'uppercase',
    letterSpacing: 0.5,
    marginBottom: spacing.sm,
  },
  card: {
    backgroundColor: colors.surface,
    borderRadius: 12,
    padding: spacing.lg,
  },
  infoRow: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    paddingVertical: spacing.xs,
  },
  infoLabel: {
    fontFamily: fontFamily.ui,
    fontSize: fontSize.md,
    color: colors.textMuted,
  },
  infoValue: {
    fontFamily: fontFamily.uiMedium,
    fontSize: fontSize.md,
    color: colors.textPrimary,
  },
  divider: {
    height: 1,
    backgroundColor: colors.border,
    marginVertical: spacing.sm,
  },
  emptyText: {
    fontFamily: fontFamily.ui,
    fontSize: fontSize.md,
    color: colors.textMuted,
    textAlign: 'center',
    paddingVertical: spacing.md,
  },
  statusSuccess: {
    color: colors.green,
  },
  statusFailed: {
    color: colors.red,
  },
  progressCard: {
    flexDirection: 'row',
    alignItems: 'center',
    backgroundColor: colors.surface,
    borderRadius: 12,
    padding: spacing.lg,
    marginBottom: spacing.lg,
  },
  progressInfo: {
    flex: 1,
    marginLeft: spacing.md,
  },
  progressLabel: {
    fontFamily: fontFamily.uiMedium,
    fontSize: fontSize.sm,
    color: colors.textPrimary,
    marginBottom: spacing.xs,
  },
  progressBarBg: {
    height: 4,
    backgroundColor: colors.border,
    borderRadius: 2,
    overflow: 'hidden',
  },
  progressBarFill: {
    height: '100%',
    backgroundColor: colors.green,
    borderRadius: 2,
  },
  progressPercent: {
    fontFamily: fontFamily.ui,
    fontSize: fontSize.xs,
    color: colors.textMuted,
    marginTop: spacing.xs,
    textAlign: 'right',
  },
  primaryButton: {
    backgroundColor: colors.green,
    borderRadius: 12,
    paddingVertical: spacing.md,
    alignItems: 'center',
    marginBottom: spacing.sm,
  },
  primaryButtonText: {
    fontFamily: fontFamily.uiSemiBold,
    fontSize: fontSize.md,
    color: colors.background,
  },
  secondaryButton: {
    backgroundColor: 'transparent',
    borderRadius: 12,
    borderWidth: 1,
    borderColor: colors.green,
    paddingVertical: spacing.md,
    alignItems: 'center',
  },
  secondaryButtonText: {
    fontFamily: fontFamily.uiSemiBold,
    fontSize: fontSize.md,
    color: colors.green,
  },
  buttonDisabled: {
    opacity: 0.5,
  },
  historyHeader: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    marginBottom: spacing.sm,
  },
  chevron: {
    fontSize: fontSize.sm,
    color: colors.textMuted,
  },
  historyItem: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    paddingVertical: spacing.xs,
  },
  historyLeft: {
    flex: 1,
  },
  historyDate: {
    fontFamily: fontFamily.uiMedium,
    fontSize: fontSize.md,
    color: colors.textPrimary,
  },
  historyMeta: {
    fontFamily: fontFamily.ui,
    fontSize: fontSize.sm,
    color: colors.textMuted,
    marginTop: 2,
  },
  historyStatus: {
    fontFamily: fontFamily.uiSemiBold,
    fontSize: fontSize.lg,
    marginLeft: spacing.md,
  },
  loader: {
    paddingVertical: spacing.md,
  },
});
