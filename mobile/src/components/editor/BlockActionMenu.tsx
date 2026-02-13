// BlockActionMenu — long press menu for block actions
import React from 'react';
import {
  View,
  Text,
  Pressable,
  StyleSheet,
  Modal,
} from 'react-native';
import { useTranslation } from 'react-i18next';
import { colors, fontFamily, fontSize, spacing } from '@/theme';

interface BlockAction {
  id: string;
  icon: string;
  labelKey: string;
  destructive?: boolean;
}

const BLOCK_ACTIONS: BlockAction[] = [
  { id: 'duplicate', icon: '📋', labelKey: 'editor.duplicate' },
  { id: 'moveUp', icon: '⬆', labelKey: 'editor.moveUp' },
  { id: 'moveDown', icon: '⬇', labelKey: 'editor.moveDown' },
  { id: 'changeType', icon: '🔄', labelKey: 'editor.changeType' },
  { id: 'delete', icon: '🗑', labelKey: 'common.delete', destructive: true },
];

interface BlockActionMenuProps {
  visible: boolean;
  onAction: (actionId: string) => void;
  onDismiss: () => void;
}

export function BlockActionMenu({ visible, onAction, onDismiss }: BlockActionMenuProps) {
  const { t } = useTranslation();

  if (!visible) return null;

  return (
    <Modal transparent animationType="fade" visible={visible} onRequestClose={onDismiss}>
      <Pressable style={styles.overlay} onPress={onDismiss}>
        <View style={styles.menu}>
          <Text style={styles.title}>{t('editor.blockActions')}</Text>
          {BLOCK_ACTIONS.map((action) => (
            <Pressable
              key={action.id}
              style={({ pressed }) => [styles.option, pressed && styles.optionPressed]}
              onPress={() => onAction(action.id)}
            >
              <Text style={styles.icon}>{action.icon}</Text>
              <Text
                style={[styles.label, action.destructive && styles.destructive]}
              >
                {t(action.labelKey)}
              </Text>
            </Pressable>
          ))}
        </View>
      </Pressable>
    </Modal>
  );
}

const styles = StyleSheet.create({
  overlay: {
    flex: 1,
    backgroundColor: 'rgba(0,0,0,0.4)',
    justifyContent: 'flex-end',
  },
  menu: {
    backgroundColor: colors.surface,
    borderTopLeftRadius: 16,
    borderTopRightRadius: 16,
    paddingBottom: 32,
    paddingTop: spacing.lg,
  },
  title: {
    fontFamily: fontFamily.uiSemiBold,
    fontSize: fontSize.md,
    color: colors.textPrimary,
    paddingHorizontal: spacing.lg,
    paddingBottom: spacing.sm,
  },
  option: {
    flexDirection: 'row',
    alignItems: 'center',
    paddingVertical: spacing.md,
    paddingHorizontal: spacing.lg,
  },
  optionPressed: {
    backgroundColor: colors.surface2,
  },
  icon: {
    fontSize: 18,
    width: 32,
    textAlign: 'center',
  },
  label: {
    fontFamily: fontFamily.uiMedium,
    fontSize: fontSize.md,
    color: colors.textPrimary,
    marginLeft: spacing.sm,
  },
  destructive: {
    color: colors.red,
  },
});
