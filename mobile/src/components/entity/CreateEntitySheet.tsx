// CreateEntitySheet — bottom sheet modal for creating a new entity
import React, { useState, useCallback, useEffect } from 'react';
import {
  View,
  Text,
  Modal,
  Pressable,
  TextInput,
  ScrollView,
  StyleSheet,
  KeyboardAvoidingView,
  Platform,
} from 'react-native';
import { useEntityStore } from '@/stores/entityStore';
import { useTranslation } from 'react-i18next';
import { colors, fontSize, fontFamily, spacing } from '@/theme';

type CreateEntitySheetProps = {
  visible: boolean;
  onDismiss: () => void;
  onCreated: () => void;
};

export function CreateEntitySheet({ visible, onDismiss, onCreated }: CreateEntitySheetProps) {
  const { t } = useTranslation();
  const { createEntity, types: existingTypes } = useEntityStore();
  const [name, setName] = useState('');
  const [type, setType] = useState('');
  const [fields, setFields] = useState<Array<{ key: string; value: string }>>([]);
  const [showTypeSuggestions, setShowTypeSuggestions] = useState(false);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [error, setError] = useState('');

  // Reset form when showing
  useEffect(() => {
    if (visible) {
      setName('');
      setType('');
      setFields([]);
      setError('');
    }
  }, [visible]);

  const filteredTypes = existingTypes.filter((tp) =>
    tp.toLowerCase().includes(type.toLowerCase()),
  );

  const handleAddField = useCallback(() => {
    setFields((prev) => [...prev, { key: '', value: '' }]);
  }, []);

  const handleRemoveField = useCallback((index: number) => {
    setFields((prev) => prev.filter((_, i) => i !== index));
  }, []);

  const handleFieldChange = useCallback(
    (index: number, key: string, value: string) => {
      setFields((prev) =>
        prev.map((f, i) => (i === index ? { key, value } : f)),
      );
    },
    [],
  );

  const handleSubmit = useCallback(async () => {
    if (!name.trim()) {
      setError(t('entity.nameRequired'));
      return;
    }
    if (!type.trim()) {
      setError(t('entity.typeRequired'));
      return;
    }

    setIsSubmitting(true);
    setError('');

    try {
      const fieldsMap: Record<string, string> = {};
      for (const f of fields) {
        if (f.key.trim()) {
          fieldsMap[f.key.trim()] = f.value;
        }
      }

      await createEntity({
        name: name.trim(),
        type: type.trim(),
        fields: Object.keys(fieldsMap).length > 0 ? fieldsMap : undefined,
      });

      onCreated();
    } catch (err) {
      setError(err instanceof Error ? err.message : t('entity.createFailed'));
    } finally {
      setIsSubmitting(false);
    }
  }, [name, type, fields, createEntity, onCreated]);

  return (
    <Modal visible={visible} transparent animationType="slide">
      <Pressable style={styles.overlay} onPress={onDismiss}>
        <KeyboardAvoidingView
          behavior={Platform.OS === 'ios' ? 'padding' : undefined}
          style={styles.keyboardView}
        >
          <Pressable style={styles.sheet} onPress={() => {}}>
            <View style={styles.handle} />
            <Text style={styles.title}>{t('entity.createNewEntity')}</Text>

            <ScrollView style={styles.form} keyboardShouldPersistTaps="handled">
              {/* Name */}
              <Text style={styles.label}>{t('entity.name')}</Text>
              <TextInput
                style={styles.input}
                value={name}
                onChangeText={setName}
                placeholder={t('entity.namePlaceholderExample')}
                placeholderTextColor={colors.textMuted}
                autoFocus
              />

              {/* Type */}
              <Text style={styles.label}>{t('entity.type')}</Text>
              <TextInput
                style={styles.input}
                value={type}
                onChangeText={(text) => {
                  setType(text);
                  setShowTypeSuggestions(text.length > 0);
                }}
                onFocus={() => setShowTypeSuggestions(type.length > 0)}
                onBlur={() => setTimeout(() => setShowTypeSuggestions(false), 200)}
                placeholder={t('entity.typePlaceholderExample')}
                placeholderTextColor={colors.textMuted}
              />
              {showTypeSuggestions && filteredTypes.length > 0 && (
                <View style={styles.suggestions}>
                  {filteredTypes.map((t) => (
                    <Pressable
                      key={t}
                      style={styles.suggestion}
                      onPress={() => {
                        setType(t);
                        setShowTypeSuggestions(false);
                      }}
                    >
                      <Text style={styles.suggestionText}>{t}</Text>
                    </Pressable>
                  ))}
                </View>
              )}

              {/* Dynamic Fields */}
              <View style={styles.fieldsHeader}>
                <Text style={styles.label}>{t('entity.fields')}</Text>
                <Pressable onPress={handleAddField}>
                  <Text style={styles.addFieldText}>+ {t('common.add')}</Text>
                </Pressable>
              </View>
              {fields.map((field, idx) => (
                <View key={idx} style={styles.fieldRow}>
                  <TextInput
                    style={[styles.input, styles.fieldKeyInput]}
                    value={field.key}
                    onChangeText={(k) => handleFieldChange(idx, k, field.value)}
                    placeholder="Key"
                    placeholderTextColor={colors.textMuted}
                  />
                  <TextInput
                    style={[styles.input, styles.fieldValueInput]}
                    value={field.value}
                    onChangeText={(v) => handleFieldChange(idx, field.key, v)}
                    placeholder="Value"
                    placeholderTextColor={colors.textMuted}
                  />
                  <Pressable onPress={() => handleRemoveField(idx)}>
                    <Text style={styles.removeBtn}>x</Text>
                  </Pressable>
                </View>
              ))}

              {error ? <Text style={styles.error}>{error}</Text> : null}
            </ScrollView>

            {/* Actions */}
            <View style={styles.actions}>
              <Pressable style={styles.cancelBtn} onPress={onDismiss}>
                <Text style={styles.cancelBtnText}>{t('common.cancel')}</Text>
              </Pressable>
              <Pressable
                style={[styles.submitBtn, isSubmitting && styles.submitBtnDisabled]}
                onPress={handleSubmit}
                disabled={isSubmitting}
              >
                <Text style={styles.submitBtnText}>
                  {isSubmitting ? t('common.saving') : t('common.save')}
                </Text>
              </Pressable>
            </View>
          </Pressable>
        </KeyboardAvoidingView>
      </Pressable>
    </Modal>
  );
}

const styles = StyleSheet.create({
  overlay: {
    flex: 1,
    backgroundColor: colors.overlay,
    justifyContent: 'flex-end',
  },
  keyboardView: {
    justifyContent: 'flex-end',
  },
  sheet: {
    backgroundColor: colors.surface,
    borderTopLeftRadius: 20,
    borderTopRightRadius: 20,
    maxHeight: '85%',
    paddingBottom: 40,
  },
  handle: {
    width: 40,
    height: 4,
    borderRadius: 2,
    backgroundColor: colors.textMuted,
    alignSelf: 'center',
    marginTop: spacing.md,
    marginBottom: spacing.lg,
  },
  title: {
    fontSize: fontSize.lg,
    fontFamily: fontFamily.uiBold,
    color: colors.textPrimary,
    textAlign: 'center',
    marginBottom: spacing.lg,
  },
  form: {
    paddingHorizontal: spacing.lg,
  },
  label: {
    fontSize: fontSize.sm,
    fontFamily: fontFamily.uiSemiBold,
    color: colors.textMuted,
    marginBottom: spacing.xs,
    marginTop: spacing.md,
  },
  input: {
    fontFamily: fontFamily.ui,
    fontSize: fontSize.md,
    color: colors.textPrimary,
    backgroundColor: colors.surface2,
    paddingHorizontal: spacing.md,
    paddingVertical: spacing.md,
    borderRadius: 10,
    borderWidth: 1,
    borderColor: colors.border,
  },
  suggestions: {
    backgroundColor: colors.surface2,
    borderRadius: 8,
    marginTop: spacing.xs,
    borderWidth: 1,
    borderColor: colors.border,
  },
  suggestion: {
    paddingHorizontal: spacing.md,
    paddingVertical: spacing.sm,
    borderBottomWidth: 1,
    borderBottomColor: colors.border,
  },
  suggestionText: {
    fontSize: fontSize.sm,
    fontFamily: fontFamily.ui,
    color: colors.textPrimary,
  },
  fieldsHeader: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    marginTop: spacing.md,
    marginBottom: spacing.xs,
  },
  addFieldText: {
    fontSize: fontSize.sm,
    fontFamily: fontFamily.uiSemiBold,
    color: colors.green,
  },
  fieldRow: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: spacing.sm,
    marginBottom: spacing.sm,
  },
  fieldKeyInput: {
    flex: 1,
  },
  fieldValueInput: {
    flex: 2,
  },
  removeBtn: {
    fontSize: fontSize.lg,
    color: colors.red,
    paddingHorizontal: spacing.sm,
  },
  error: {
    fontSize: fontSize.sm,
    fontFamily: fontFamily.ui,
    color: colors.red,
    marginTop: spacing.md,
  },
  actions: {
    flexDirection: 'row',
    gap: spacing.md,
    paddingHorizontal: spacing.lg,
    paddingTop: spacing.lg,
  },
  cancelBtn: {
    flex: 1,
    paddingVertical: spacing.md,
    borderRadius: 10,
    backgroundColor: colors.surface2,
    alignItems: 'center',
  },
  cancelBtnText: {
    fontSize: fontSize.md,
    fontFamily: fontFamily.uiSemiBold,
    color: colors.textMuted,
  },
  submitBtn: {
    flex: 2,
    paddingVertical: spacing.md,
    borderRadius: 10,
    backgroundColor: colors.green,
    alignItems: 'center',
  },
  submitBtnDisabled: {
    opacity: 0.6,
  },
  submitBtnText: {
    fontSize: fontSize.md,
    fontFamily: fontFamily.uiBold,
    color: colors.background,
  },
});
