// SlashMenu — block type selector triggered by "/"
import React, { useMemo } from 'react';
import {
  View,
  Text,
  FlatList,
  Pressable,
  StyleSheet,
  Modal,
} from 'react-native';
import { useTranslation } from 'react-i18next';
import { colors, fontFamily, fontSize, spacing } from '@/theme';
import { BLOCK_OPTIONS } from './types';
import type { BlockType } from '@/types/chat';

interface SlashMenuProps {
  visible: boolean;
  filter: string;
  onSelect: (type: BlockType) => void;
  onDismiss: () => void;
}

export function SlashMenu({ visible, filter, onSelect, onDismiss }: SlashMenuProps) {
  const { t } = useTranslation();

  const filtered = useMemo(
    () =>
      BLOCK_OPTIONS.filter(
        (opt) =>
          filter === '' ||
          t(opt.labelKey).toLowerCase().includes(filter.toLowerCase()) ||
          opt.type.toLowerCase().includes(filter.toLowerCase()),
      ),
    [filter, t],
  );

  if (!visible) return null;

  return (
    <Modal transparent animationType="fade" visible={visible} onRequestClose={onDismiss}>
      <Pressable style={styles.overlay} onPress={onDismiss}>
        <View style={styles.menu}>
          <Text style={styles.title}>{t('editor.addBlock')}</Text>
          <FlatList
            data={filtered}
            keyExtractor={(item) => item.type}
            keyboardShouldPersistTaps="handled"
            renderItem={({ item }) => (
              <Pressable
                style={({ pressed }) => [styles.option, pressed && styles.optionPressed]}
                onPress={() => onSelect(item.type)}
              >
                <Text style={styles.icon}>{item.icon}</Text>
                <View style={styles.optionText}>
                  <Text style={styles.label}>{t(item.labelKey)}</Text>
                  <Text style={styles.description}>{t(item.descriptionKey)}</Text>
                </View>
              </Pressable>
            )}
            ListEmptyComponent={
              <Text style={styles.empty}>{t('common.notFound')}</Text>
            }
            style={styles.list}
          />
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
    maxHeight: '60%',
    paddingBottom: 24,
  },
  title: {
    fontFamily: fontFamily.uiSemiBold,
    fontSize: fontSize.md,
    color: colors.textPrimary,
    paddingHorizontal: spacing.lg,
    paddingTop: spacing.lg,
    paddingBottom: spacing.sm,
  },
  list: {
    paddingHorizontal: spacing.sm,
  },
  option: {
    flexDirection: 'row',
    alignItems: 'center',
    paddingVertical: spacing.md,
    paddingHorizontal: spacing.md,
    borderRadius: 8,
  },
  optionPressed: {
    backgroundColor: colors.surface2,
  },
  icon: {
    fontSize: 20,
    width: 36,
    textAlign: 'center',
  },
  optionText: {
    flex: 1,
    marginLeft: spacing.sm,
  },
  label: {
    fontFamily: fontFamily.uiMedium,
    fontSize: fontSize.md,
    color: colors.textPrimary,
  },
  description: {
    fontFamily: fontFamily.ui,
    fontSize: fontSize.xs,
    color: colors.textMuted,
    marginTop: 2,
  },
  empty: {
    fontFamily: fontFamily.ui,
    fontSize: fontSize.sm,
    color: colors.textMuted,
    textAlign: 'center',
    paddingVertical: spacing.xl,
  },
});
