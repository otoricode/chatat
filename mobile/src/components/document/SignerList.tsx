// Signer list component — shows signers and their status
import React from 'react';
import { View, Text, Pressable, StyleSheet, FlatList } from 'react-native';
import { colors, fontSize, fontFamily, spacing } from '@/theme';
import { useTranslation } from 'react-i18next';
import type { DocumentSigner } from '@/types/chat';

interface SignerListProps {
  signers: DocumentSigner[];
  isOwner: boolean;
  onRemove?: (userId: string) => void;
}

export function SignerList({ signers, isOwner, onRemove }: SignerListProps) {
  const { t } = useTranslation();
  if (signers.length === 0) {
    return (
      <View style={styles.empty}>
        <Text style={styles.emptyText}>{t('document.pendingSignatures')}</Text>
      </View>
    );
  }

  return (
    <FlatList
      data={signers}
      keyExtractor={(item) => item.userId}
      renderItem={({ item }) => (
        <View style={styles.row}>
          <View style={styles.avatar}>
            <Text style={styles.avatarText}>
              {item.signerName ? item.signerName.charAt(0).toUpperCase() : '?'}
            </Text>
          </View>
          <View style={styles.info}>
            <Text style={styles.name}>
              {item.signerName || 'Penandatangan'}
            </Text>
            <Text style={[styles.status, item.signedAt ? styles.signed : styles.pending]}>
              {item.signedAt ? t('document.signed') : t('document.pendingSignatures')}
            </Text>
          </View>
          {item.signedAt && <Text style={styles.checkIcon}>✅</Text>}
          {!item.signedAt && isOwner && onRemove && (
            <Pressable onPress={() => onRemove(item.userId)} style={styles.removeBtn}>
              <Text style={styles.removeText}>{t('common.remove')}</Text>
            </Pressable>
          )}
        </View>
      )}
      scrollEnabled={false}
    />
  );
}

const styles = StyleSheet.create({
  empty: {
    paddingVertical: spacing.lg,
    alignItems: 'center',
  },
  emptyText: {
    fontFamily: fontFamily.ui,
    fontSize: fontSize.sm,
    color: colors.textMuted,
  },
  row: {
    flexDirection: 'row',
    alignItems: 'center',
    paddingVertical: spacing.sm,
    paddingHorizontal: spacing.md,
    gap: spacing.md,
  },
  avatar: {
    width: 36,
    height: 36,
    borderRadius: 18,
    backgroundColor: colors.surface2,
    justifyContent: 'center',
    alignItems: 'center',
  },
  avatarText: {
    fontFamily: fontFamily.uiSemiBold,
    fontSize: fontSize.md,
    color: colors.textPrimary,
  },
  info: {
    flex: 1,
  },
  name: {
    fontFamily: fontFamily.uiMedium,
    fontSize: fontSize.md,
    color: colors.textPrimary,
  },
  status: {
    fontFamily: fontFamily.ui,
    fontSize: fontSize.xs,
    marginTop: 2,
  },
  signed: {
    color: colors.green,
  },
  pending: {
    color: colors.textMuted,
  },
  checkIcon: {
    fontSize: 16,
  },
  removeBtn: {
    paddingHorizontal: spacing.sm,
    paddingVertical: spacing.xs,
  },
  removeText: {
    fontFamily: fontFamily.uiMedium,
    fontSize: fontSize.sm,
    color: colors.red,
  },
});
