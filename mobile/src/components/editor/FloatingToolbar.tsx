// FloatingToolbar — text formatting toolbar
import React from 'react';
import { View, Text, Pressable, StyleSheet, Modal } from 'react-native';
import { colors, fontFamily, fontSize, spacing } from '@/theme';

interface ToolbarAction {
  id: string;
  label: string;
  display: string;
}

const TOOLBAR_ACTIONS: ToolbarAction[] = [
  { id: 'bold', label: 'Bold', display: 'B' },
  { id: 'italic', label: 'Italic', display: 'I' },
  { id: 'underline', label: 'Underline', display: 'U' },
  { id: 'strikethrough', label: 'Coret', display: 'S' },
  { id: 'code', label: 'Kode', display: '</>' },
];

interface FloatingToolbarProps {
  visible: boolean;
  onAction: (format: string) => void;
  onDismiss: () => void;
}

export function FloatingToolbar({ visible, onAction, onDismiss }: FloatingToolbarProps) {
  if (!visible) return null;

  return (
    <Modal transparent animationType="fade" visible={visible} onRequestClose={onDismiss}>
      <Pressable style={styles.overlay} onPress={onDismiss}>
        <View style={styles.toolbar}>
          {TOOLBAR_ACTIONS.map((action) => (
            <Pressable
              key={action.id}
              style={({ pressed }) => [styles.button, pressed && styles.buttonPressed]}
              onPress={() => onAction(action.id)}
            >
              <Text
                style={[
                  styles.buttonText,
                  action.id === 'bold' && styles.bold,
                  action.id === 'italic' && styles.italic,
                  action.id === 'strikethrough' && styles.strike,
                ]}
              >
                {action.display}
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
    justifyContent: 'center',
    alignItems: 'center',
  },
  toolbar: {
    flexDirection: 'row',
    backgroundColor: colors.surface2,
    borderRadius: 8,
    paddingVertical: spacing.xs,
    paddingHorizontal: spacing.xs,
    shadowColor: colors.black,
    shadowOffset: { width: 0, height: 2 },
    shadowOpacity: 0.3,
    shadowRadius: 8,
    elevation: 8,
  },
  button: {
    paddingHorizontal: spacing.md,
    paddingVertical: spacing.sm,
    borderRadius: 6,
    minWidth: 40,
    alignItems: 'center',
  },
  buttonPressed: {
    backgroundColor: colors.surface,
  },
  buttonText: {
    fontFamily: fontFamily.document,
    fontSize: fontSize.md,
    color: colors.textPrimary,
  },
  bold: {
    fontFamily: fontFamily.documentBold,
  },
  italic: {
    fontStyle: 'italic',
  },
  strike: {
    textDecorationLine: 'line-through',
  },
});
