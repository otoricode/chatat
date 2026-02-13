// Entity detail screen — view entity info, fields, and linked documents
import React, { useCallback, useEffect, useState } from 'react';
import {
  View,
  Text,
  FlatList,
  Pressable,
  Alert,
  StyleSheet,
  ScrollView,
} from 'react-native';
import { SafeAreaView } from 'react-native-safe-area-context';
import type { NativeStackScreenProps } from '@react-navigation/native-stack';
import type { DocumentStackParamList } from '@/navigation/types';
import { entitiesApi } from '@/services/api/entities';
import { useEntityStore } from '@/stores/entityStore';
import { colors, fontSize, fontFamily, spacing } from '@/theme';
import type { Entity, Document } from '@/types/chat';
import { ConfirmDialog } from '@/components/shared/ConfirmDialog';

type Props = NativeStackScreenProps<DocumentStackParamList, 'EntityDetail'>;

export function EntityDetailScreen({ route, navigation }: Props) {
  const { entityId } = route.params;
  const { deleteEntity } = useEntityStore();
  const [entity, setEntity] = useState<Entity | null>(null);
  const [documents, setDocuments] = useState<Document[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [showDelete, setShowDelete] = useState(false);
  const [isEditing, setIsEditing] = useState(false);
  const [editName, setEditName] = useState('');
  const [editType, setEditType] = useState('');
  const [editFields, setEditFields] = useState<Record<string, string>>({});

  const loadData = useCallback(async () => {
    setIsLoading(true);
    try {
      const [entityRes, docsRes] = await Promise.all([
        entitiesApi.getById(entityId),
        entitiesApi.listDocuments(entityId),
      ]);
      const entityData = (entityRes.data as unknown as { data: Entity }).data;
      const docsData = (docsRes.data as unknown as { data: Document[] }).data;
      setEntity(entityData);
      setDocuments(docsData ?? []);
      setEditName(entityData.name);
      setEditType(entityData.type);
      setEditFields(entityData.fields ?? {});
    } catch {
      Alert.alert('Gagal memuat entity');
    } finally {
      setIsLoading(false);
    }
  }, [entityId]);

  useEffect(() => {
    loadData();
  }, [loadData]);

  const handleDelete = useCallback(async () => {
    try {
      await deleteEntity(entityId);
      navigation.goBack();
    } catch {
      Alert.alert('Gagal menghapus entity');
    }
  }, [deleteEntity, entityId, navigation]);

  const handleSave = useCallback(async () => {
    try {
      const res = await entitiesApi.update(entityId, {
        name: editName,
        type: editType,
        fields: editFields,
      });
      const updated = (res.data as unknown as { data: Entity }).data;
      setEntity(updated);
      setIsEditing(false);
    } catch {
      Alert.alert('Gagal menyimpan perubahan');
    }
  }, [entityId, editName, editType, editFields]);

  const handleAddField = useCallback(() => {
    setEditFields((prev) => ({ ...prev, '': '' }));
  }, []);

  const handleDocPress = useCallback(
    (doc: Document) => {
      navigation.navigate('DocumentEditor', { documentId: doc.id });
    },
    [navigation],
  );

  if (isLoading || !entity) {
    return (
      <SafeAreaView style={styles.container}>
        <View style={styles.centered}>
          <Text style={styles.loadingText}>Memuat...</Text>
        </View>
      </SafeAreaView>
    );
  }

  return (
    <SafeAreaView style={styles.container} edges={['top']}>
      <ScrollView contentContainerStyle={styles.scrollContent}>
        {/* Header */}
        <View style={styles.header}>
          <View style={[styles.avatar, { backgroundColor: getTypeColor(entity.type) }]}>
            <Text style={styles.avatarText}>{entity.name.charAt(0)}</Text>
          </View>
          {isEditing ? (
            <View style={styles.editNameContainer}>
              <EditableInput
                value={editName}
                onChangeText={setEditName}
                placeholder="Nama entity"
                style={styles.editNameInput}
              />
              <EditableInput
                value={editType}
                onChangeText={setEditType}
                placeholder="Tipe"
                style={styles.editTypeInput}
              />
            </View>
          ) : (
            <View>
              <Text style={styles.name}>{entity.name}</Text>
              <View style={[styles.typePill, { backgroundColor: getTypeColor(entity.type) + '20' }]}>
                <Text style={[styles.typeText, { color: getTypeColor(entity.type) }]}>{entity.type}</Text>
              </View>
            </View>
          )}
        </View>

        {/* Actions */}
        <View style={styles.actions}>
          {isEditing ? (
            <>
              <Pressable style={styles.actionBtn} onPress={handleSave}>
                <Text style={styles.actionBtnText}>Simpan</Text>
              </Pressable>
              <Pressable style={styles.actionBtn} onPress={() => setIsEditing(false)}>
                <Text style={[styles.actionBtnText, { color: colors.textMuted }]}>Batal</Text>
              </Pressable>
            </>
          ) : (
            <>
              <Pressable style={styles.actionBtn} onPress={() => setIsEditing(true)}>
                <Text style={styles.actionBtnText}>Edit</Text>
              </Pressable>
              <Pressable style={styles.actionBtn} onPress={() => setShowDelete(true)}>
                <Text style={[styles.actionBtnText, { color: colors.red }]}>Hapus</Text>
              </Pressable>
            </>
          )}
        </View>

        {/* Fields */}
        <View style={styles.section}>
          <Text style={styles.sectionTitle}>Fields</Text>
          {isEditing ? (
            <>
              {Object.entries(editFields).map(([key, value], idx) => (
                <View key={idx} style={styles.fieldRow}>
                  <EditableInput
                    value={key}
                    onChangeText={(newKey) => {
                      setEditFields((prev) => {
                        const entries = Object.entries(prev);
                        entries[idx] = [newKey, value];
                        return Object.fromEntries(entries);
                      });
                    }}
                    placeholder="Key"
                    style={styles.fieldKeyInput}
                  />
                  <EditableInput
                    value={value}
                    onChangeText={(newVal) => {
                      setEditFields((prev) => ({ ...prev, [key]: newVal }));
                    }}
                    placeholder="Value"
                    style={styles.fieldValueInput}
                  />
                  <Pressable
                    onPress={() => {
                      setEditFields((prev) => {
                        const copy = { ...prev };
                        delete copy[key];
                        return copy;
                      });
                    }}
                  >
                    <Text style={styles.removeField}>x</Text>
                  </Pressable>
                </View>
              ))}
              <Pressable style={styles.addFieldBtn} onPress={handleAddField}>
                <Text style={styles.addFieldText}>+ Tambah Field</Text>
              </Pressable>
            </>
          ) : (
            <>
              {Object.keys(entity.fields ?? {}).length === 0 ? (
                <Text style={styles.emptyText}>Belum ada field</Text>
              ) : (
                Object.entries(entity.fields).map(([key, value]) => (
                  <View key={key} style={styles.fieldDisplay}>
                    <Text style={styles.fieldKey}>{key}</Text>
                    <Text style={styles.fieldValue}>{value}</Text>
                  </View>
                ))
              )}
            </>
          )}
        </View>

        {/* Linked Documents */}
        <View style={styles.section}>
          <Text style={styles.sectionTitle}>
            Dokumen Terkait ({documents.length})
          </Text>
          {documents.length === 0 ? (
            <Text style={styles.emptyText}>Belum ada dokumen terkait</Text>
          ) : (
            documents.map((doc) => (
              <Pressable
                key={doc.id}
                style={({ pressed }) => [styles.docItem, pressed && { opacity: 0.7 }]}
                onPress={() => handleDocPress(doc)}
              >
                <Text style={styles.docIcon}>{doc.icon || '📄'}</Text>
                <Text style={styles.docTitle} numberOfLines={1}>
                  {doc.title || 'Tanpa Judul'}
                </Text>
              </Pressable>
            ))
          )}
        </View>

        {entity.contactUserId && (
          <View style={styles.section}>
            <Text style={styles.sectionTitle}>Dibuat dari Kontak</Text>
            <Text style={styles.contactInfo}>Terhubung dengan kontak pengguna</Text>
          </View>
        )}
      </ScrollView>

      <ConfirmDialog
        visible={showDelete}
        title="Hapus Entity"
        message={`Yakin ingin menghapus "${entity.name}"? Entity akan dilepas dari semua dokumen terkait.`}
        confirmText="Hapus"
        variant="danger"
        onConfirm={handleDelete}
        onCancel={() => setShowDelete(false)}
      />
    </SafeAreaView>
  );
}

// Simple inline text input component
function EditableInput({
  value,
  onChangeText,
  placeholder,
  style,
}: {
  value: string;
  onChangeText: (text: string) => void;
  placeholder: string;
  style?: object;
}) {
  const { TextInput } = require('react-native');
  return (
    <TextInput
      value={value}
      onChangeText={onChangeText}
      placeholder={placeholder}
      placeholderTextColor={colors.textMuted}
      style={[styles.input, style]}
    />
  );
}

const TYPE_COLORS: Record<string, string> = {
  Orang: colors.blue,
  Lahan: colors.green,
  Aset: colors.yellow,
  Proyek: colors.purple,
  Lokasi: colors.red,
};

function getTypeColor(type: string): string {
  return TYPE_COLORS[type] ?? colors.purple;
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: colors.background,
  },
  centered: {
    flex: 1,
    justifyContent: 'center',
    alignItems: 'center',
  },
  loadingText: {
    color: colors.textMuted,
    fontFamily: fontFamily.ui,
    fontSize: fontSize.md,
  },
  scrollContent: {
    padding: spacing.lg,
    paddingBottom: 40,
  },
  header: {
    alignItems: 'center',
    marginBottom: spacing.xl,
  },
  avatar: {
    width: 64,
    height: 64,
    borderRadius: 32,
    justifyContent: 'center',
    alignItems: 'center',
    marginBottom: spacing.md,
  },
  avatarText: {
    fontSize: fontSize.xxl,
    fontFamily: fontFamily.uiBold,
    color: colors.background,
  },
  name: {
    fontSize: fontSize.xl,
    fontFamily: fontFamily.uiBold,
    color: colors.textPrimary,
    textAlign: 'center',
  },
  typePill: {
    paddingHorizontal: spacing.md,
    paddingVertical: spacing.xs,
    borderRadius: 12,
    alignSelf: 'center',
    marginTop: spacing.sm,
  },
  typeText: {
    fontSize: fontSize.sm,
    fontFamily: fontFamily.uiSemiBold,
  },
  editNameContainer: {
    width: '100%',
    gap: spacing.sm,
  },
  editNameInput: {
    fontSize: fontSize.xl,
    textAlign: 'center',
  },
  editTypeInput: {
    fontSize: fontSize.md,
    textAlign: 'center',
  },
  actions: {
    flexDirection: 'row',
    justifyContent: 'center',
    gap: spacing.md,
    marginBottom: spacing.xl,
  },
  actionBtn: {
    paddingHorizontal: spacing.lg,
    paddingVertical: spacing.sm,
    borderRadius: 8,
    backgroundColor: colors.surface2,
  },
  actionBtnText: {
    fontSize: fontSize.sm,
    fontFamily: fontFamily.uiSemiBold,
    color: colors.green,
  },
  section: {
    marginBottom: spacing.xl,
  },
  sectionTitle: {
    fontSize: fontSize.md,
    fontFamily: fontFamily.uiBold,
    color: colors.textPrimary,
    marginBottom: spacing.md,
  },
  emptyText: {
    fontSize: fontSize.sm,
    fontFamily: fontFamily.ui,
    color: colors.textMuted,
  },
  fieldDisplay: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    paddingVertical: spacing.sm,
    borderBottomWidth: 1,
    borderBottomColor: colors.border,
  },
  fieldKey: {
    fontSize: fontSize.sm,
    fontFamily: fontFamily.uiSemiBold,
    color: colors.textMuted,
    flex: 1,
  },
  fieldValue: {
    fontSize: fontSize.sm,
    fontFamily: fontFamily.ui,
    color: colors.textPrimary,
    flex: 2,
    textAlign: 'right',
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
  removeField: {
    fontSize: fontSize.lg,
    color: colors.red,
    paddingHorizontal: spacing.sm,
  },
  addFieldBtn: {
    paddingVertical: spacing.sm,
  },
  addFieldText: {
    fontSize: fontSize.sm,
    fontFamily: fontFamily.uiSemiBold,
    color: colors.green,
  },
  input: {
    fontFamily: fontFamily.ui,
    fontSize: fontSize.sm,
    color: colors.textPrimary,
    backgroundColor: colors.surface,
    paddingHorizontal: spacing.md,
    paddingVertical: spacing.sm,
    borderRadius: 8,
    borderWidth: 1,
    borderColor: colors.border,
  },
  docItem: {
    flexDirection: 'row',
    alignItems: 'center',
    paddingVertical: spacing.md,
    borderBottomWidth: 1,
    borderBottomColor: colors.border,
  },
  docIcon: {
    fontSize: fontSize.lg,
    marginRight: spacing.md,
  },
  docTitle: {
    fontSize: fontSize.md,
    fontFamily: fontFamily.ui,
    color: colors.textPrimary,
    flex: 1,
  },
  contactInfo: {
    fontSize: fontSize.sm,
    fontFamily: fontFamily.ui,
    color: colors.textMuted,
  },
});
