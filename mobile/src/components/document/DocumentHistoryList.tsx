// Document history timeline component
import React from 'react';
import { View, Text, StyleSheet, FlatList } from 'react-native';
import { useTranslation } from 'react-i18next';
import type { TFunction } from 'i18next';
import { colors, fontSize, fontFamily, spacing } from '@/theme';
import type { DocumentHistory } from '@/types/chat';

interface DocumentHistoryListProps {
  history: DocumentHistory[];
}

const ACTION_ICONS: Record<string, string> = {
  created: '📄',
  updated: '✏️',
  locked_manual: '🔒',
  locked_signatures: '✍️',
  unlocked: '🔓',
  signed: '✅',
  collaborator_added: '👤',
  collaborator_removed: '👤',
  signer_added: '✍️',
  signer_removed: '✍️',
  duplicated: '📋',
};

function formatDate(dateStr: string, t: TFunction): string {
  const date = new Date(dateStr);
  const now = new Date();
  const diffMs = now.getTime() - date.getTime();
  const diffMin = Math.floor(diffMs / 60000);
  const diffHr = Math.floor(diffMin / 60);
  const diffDay = Math.floor(diffHr / 24);

  if (diffMin < 1) return t('time.justNow');
  if (diffMin < 60) return t('time.minutesAgo', { count: diffMin });
  if (diffHr < 24) return t('time.hoursAgo', { count: diffHr });
  if (diffDay < 7) return t('time.daysAgo', { count: diffDay });

  return date.toLocaleDateString('id-ID', {
    day: 'numeric',
    month: 'short',
    year: 'numeric',
  });
}

export function DocumentHistoryList({ history }: DocumentHistoryListProps) {
  const { t } = useTranslation();

  if (history.length === 0) {
    return (
      <View style={styles.empty}>
        <Text style={styles.emptyText}>{t('document.noActivity')}</Text>
      </View>
    );
  }

  return (
    <FlatList
      data={history}
      keyExtractor={(item) => item.id}
      renderItem={({ item, index }) => (
        <View style={styles.row}>
          {/* Timeline line */}
          <View style={styles.timeline}>
            {index > 0 && <View style={styles.lineTop} />}
            <View style={styles.dot}>
              <Text style={styles.dotIcon}>
                {ACTION_ICONS[item.action] || '📋'}
              </Text>
            </View>
            {index < history.length - 1 && <View style={styles.lineBottom} />}
          </View>

          {/* Content */}
          <View style={styles.content}>
            <Text style={styles.details}>{item.details}</Text>
            <Text style={styles.time}>{formatDate(item.createdAt, t)}</Text>
          </View>
        </View>
      )}
      scrollEnabled={false}
    />
  );
}

const styles = StyleSheet.create({
  empty: {
    paddingVertical: spacing.xl,
    alignItems: 'center',
  },
  emptyText: {
    fontFamily: fontFamily.ui,
    fontSize: fontSize.sm,
    color: colors.textMuted,
  },
  row: {
    flexDirection: 'row',
    minHeight: 48,
  },
  timeline: {
    width: 40,
    alignItems: 'center',
  },
  lineTop: {
    width: 2,
    flex: 1,
    backgroundColor: colors.border,
  },
  lineBottom: {
    width: 2,
    flex: 1,
    backgroundColor: colors.border,
  },
  dot: {
    width: 28,
    height: 28,
    borderRadius: 14,
    backgroundColor: colors.surface2,
    justifyContent: 'center',
    alignItems: 'center',
  },
  dotIcon: {
    fontSize: 14,
  },
  content: {
    flex: 1,
    paddingVertical: spacing.sm,
    paddingLeft: spacing.sm,
  },
  details: {
    fontFamily: fontFamily.ui,
    fontSize: fontSize.sm,
    color: colors.textPrimary,
    lineHeight: 18,
  },
  time: {
    fontFamily: fontFamily.ui,
    fontSize: fontSize.xs,
    color: colors.textMuted,
    marginTop: 2,
  },
});
