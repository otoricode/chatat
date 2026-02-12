// ConfirmDialog — confirmation modal
import React from 'react';
import { View, Text, Modal, StyleSheet, Pressable } from 'react-native';
import { colors, fontSize, spacing, fontFamily } from '@/theme';

type ConfirmDialogProps = {
  visible: boolean;
  title: string;
  message: string;
  confirmText?: string;
  cancelText?: string;
  variant?: 'default' | 'danger';
  onConfirm: () => void;
  onCancel: () => void;
};

export function ConfirmDialog({
  visible,
  title,
  message,
  confirmText = 'Ya',
  cancelText = 'Batal',
  variant = 'default',
  onConfirm,
  onCancel,
}: ConfirmDialogProps) {
  return (
    <Modal visible={visible} transparent animationType="fade">
      <Pressable style={styles.overlay} onPress={onCancel}>
        <View style={styles.dialog}>
          <Text style={styles.title}>{title}</Text>
          <Text style={styles.message}>{message}</Text>
          <View style={styles.buttons}>
            <Pressable style={styles.cancelButton} onPress={onCancel}>
              <Text style={styles.cancelText}>{cancelText}</Text>
            </Pressable>
            <Pressable
              style={[
                styles.confirmButton,
                variant === 'danger' && styles.dangerButton,
              ]}
              onPress={onConfirm}
            >
              <Text
                style={[
                  styles.confirmText,
                  variant === 'danger' && styles.dangerText,
                ]}
              >
                {confirmText}
              </Text>
            </Pressable>
          </View>
        </View>
      </Pressable>
    </Modal>
  );
}

const styles = StyleSheet.create({
  overlay: {
    flex: 1,
    backgroundColor: colors.overlay,
    justifyContent: 'center',
    alignItems: 'center',
    padding: spacing.xxl,
  },
  dialog: {
    backgroundColor: colors.surface,
    borderRadius: 16,
    padding: spacing.xxl,
    width: '100%',
    maxWidth: 340,
  },
  title: {
    fontFamily: fontFamily.uiBold,
    fontSize: fontSize.lg,
    color: colors.textPrimary,
    marginBottom: spacing.sm,
  },
  message: {
    fontFamily: fontFamily.ui,
    fontSize: fontSize.sm,
    color: colors.textMuted,
    lineHeight: 20,
    marginBottom: spacing.xxl,
  },
  buttons: {
    flexDirection: 'row',
    justifyContent: 'flex-end',
    gap: spacing.md,
  },
  cancelButton: {
    paddingVertical: spacing.sm,
    paddingHorizontal: spacing.lg,
  },
  cancelText: {
    fontFamily: fontFamily.uiMedium,
    fontSize: fontSize.sm,
    color: colors.textMuted,
  },
  confirmButton: {
    paddingVertical: spacing.sm,
    paddingHorizontal: spacing.lg,
    backgroundColor: colors.green,
    borderRadius: 8,
  },
  dangerButton: {
    backgroundColor: colors.red,
  },
  confirmText: {
    fontFamily: fontFamily.uiSemiBold,
    fontSize: fontSize.sm,
    color: colors.background,
  },
  dangerText: {
    color: colors.white,
  },
});
