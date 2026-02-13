// Lock action sheet — lets document owner manage lock state
import React from 'react';
import {
  View,
  Text,
  Pressable,
  Modal,
  StyleSheet,
  ActivityIndicator,
} from 'react-native';
import { colors, fontSize, fontFamily, spacing } from '@/theme';

interface LockActionSheetProps {
  visible: boolean;
  locked: boolean;
  lockedBy?: string | null;
  loading?: boolean;
  onLockManual: () => void;
  onLockSignatures: () => void;
  onUnlock: () => void;
  onClose: () => void;
}

export function LockActionSheet({
  visible,
  locked,
  lockedBy,
  loading,
  onLockManual,
  onLockSignatures,
  onUnlock,
  onClose,
}: LockActionSheetProps) {
  return (
    <Modal
      visible={visible}
      transparent
      animationType="slide"
      onRequestClose={onClose}
    >
      <Pressable style={styles.overlay} onPress={onClose}>
        <View style={styles.sheet}>
          <View style={styles.handle} />
          <Text style={styles.title}>
            {locked ? 'Dokumen Terkunci' : 'Kunci Dokumen'}
          </Text>

          {loading && (
            <ActivityIndicator
              size="small"
              color={colors.green}
              style={{ marginVertical: spacing.md }}
            />
          )}

          {!loading && !locked && (
            <>
              <Pressable style={styles.option} onPress={onLockManual}>
                <Text style={styles.optionIcon}>🔒</Text>
                <View style={styles.optionContent}>
                  <Text style={styles.optionLabel}>Kunci Manual</Text>
                  <Text style={styles.optionDesc}>
                    Kunci dokumen agar tidak bisa diubah
                  </Text>
                </View>
              </Pressable>

              <Pressable style={styles.option} onPress={onLockSignatures}>
                <Text style={styles.optionIcon}>✍️</Text>
                <View style={styles.optionContent}>
                  <Text style={styles.optionLabel}>Kunci dengan Tanda Tangan</Text>
                  <Text style={styles.optionDesc}>
                    Kunci setelah semua penandatangan menandatangani
                  </Text>
                </View>
              </Pressable>
            </>
          )}

          {!loading && locked && (
            <>
              <View style={styles.statusRow}>
                <Text style={styles.statusIcon}>
                  {lockedBy === 'signatures' ? '✅' : '🔒'}
                </Text>
                <Text style={styles.statusText}>
                  {lockedBy === 'signatures'
                    ? 'Dokumen dikunci dengan tanda tangan'
                    : 'Dokumen dikunci secara manual'}
                </Text>
              </View>

              {lockedBy === 'manual' && (
                <Pressable style={[styles.option, styles.dangerOption]} onPress={onUnlock}>
                  <Text style={styles.optionIcon}>🔓</Text>
                  <View style={styles.optionContent}>
                    <Text style={[styles.optionLabel, { color: colors.red }]}>
                      Buka Kunci
                    </Text>
                    <Text style={styles.optionDesc}>
                      Dokumen dapat diubah kembali
                    </Text>
                  </View>
                </Pressable>
              )}
            </>
          )}

          <Pressable style={styles.cancelBtn} onPress={onClose}>
            <Text style={styles.cancelText}>Batal</Text>
          </Pressable>
        </View>
      </Pressable>
    </Modal>
  );
}

const styles = StyleSheet.create({
  overlay: {
    flex: 1,
    backgroundColor: 'rgba(0, 0, 0, 0.5)',
    justifyContent: 'flex-end',
  },
  sheet: {
    backgroundColor: colors.surface,
    borderTopLeftRadius: 16,
    borderTopRightRadius: 16,
    paddingHorizontal: spacing.lg,
    paddingBottom: spacing.xxxl,
  },
  handle: {
    width: 36,
    height: 4,
    backgroundColor: colors.border,
    borderRadius: 2,
    alignSelf: 'center',
    marginTop: spacing.sm,
    marginBottom: spacing.lg,
  },
  title: {
    fontFamily: fontFamily.uiSemiBold,
    fontSize: fontSize.lg,
    color: colors.textPrimary,
    marginBottom: spacing.lg,
  },
  option: {
    flexDirection: 'row',
    alignItems: 'center',
    paddingVertical: spacing.md,
    paddingHorizontal: spacing.sm,
    borderRadius: 8,
    gap: spacing.md,
  },
  dangerOption: {
    marginTop: spacing.sm,
  },
  optionIcon: {
    fontSize: 24,
  },
  optionContent: {
    flex: 1,
  },
  optionLabel: {
    fontFamily: fontFamily.uiMedium,
    fontSize: fontSize.md,
    color: colors.textPrimary,
  },
  optionDesc: {
    fontFamily: fontFamily.ui,
    fontSize: fontSize.sm,
    color: colors.textMuted,
    marginTop: 2,
  },
  statusRow: {
    flexDirection: 'row',
    alignItems: 'center',
    backgroundColor: colors.surface2,
    paddingHorizontal: spacing.md,
    paddingVertical: spacing.md,
    borderRadius: 8,
    gap: spacing.sm,
    marginBottom: spacing.md,
  },
  statusIcon: {
    fontSize: 20,
  },
  statusText: {
    fontFamily: fontFamily.ui,
    fontSize: fontSize.md,
    color: colors.textPrimary,
    flex: 1,
  },
  cancelBtn: {
    alignItems: 'center',
    paddingVertical: spacing.md,
    marginTop: spacing.md,
  },
  cancelText: {
    fontFamily: fontFamily.uiMedium,
    fontSize: fontSize.md,
    color: colors.textMuted,
  },
});
