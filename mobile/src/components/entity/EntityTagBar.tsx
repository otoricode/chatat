// EntityTagBar — displays entities linked to a document in the editor
import React, { useCallback, useEffect, useState } from 'react';
import {
  View,
  Text,
  ScrollView,
  Pressable,
  Modal,
  TextInput,
  FlatList,
  Alert,
  StyleSheet,
} from 'react-native';
import { entitiesApi } from '@/services/api/entities';
import { useEntityStore } from '@/stores/entityStore';
import { colors, fontSize, fontFamily, spacing } from '@/theme';
import type { Entity } from '@/types/chat';
import { ConfirmDialog } from '@/components/shared/ConfirmDialog';

type EntityTagBarProps = {
  documentId: string;
  onEntityPress?: (entityId: string) => void;
};

export function EntityTagBar({ documentId, onEntityPress }: EntityTagBarProps) {
  const [entities, setEntities] = useState<Entity[]>([]);
  const [showSelector, setShowSelector] = useState(false);
  const [unlinkTarget, setUnlinkTarget] = useState<Entity | null>(null);

  const loadEntities = useCallback(async () => {
    try {
      const res = await entitiesApi.getDocumentEntities(documentId);
      const data = (res.data as unknown as { data: Entity[] }).data;
      setEntities(data ?? []);
    } catch {
      // silently ignore
    }
  }, [documentId]);

  useEffect(() => {
    loadEntities();
  }, [loadEntities]);

  const handleLink = useCallback(
    async (entityId: string) => {
      try {
        await entitiesApi.linkToDocument(documentId, entityId);
        setShowSelector(false);
        loadEntities();
      } catch {
        Alert.alert('Gagal menautkan entity');
      }
    },
    [documentId, loadEntities],
  );

  const handleUnlink = useCallback(async () => {
    if (!unlinkTarget) return;
    try {
      await entitiesApi.unlinkFromDocument(documentId, unlinkTarget.id);
      setUnlinkTarget(null);
      loadEntities();
    } catch {
      Alert.alert('Gagal melepas entity');
    }
  }, [documentId, unlinkTarget, loadEntities]);

  return (
    <View style={styles.container}>
      <ScrollView
        horizontal
        showsHorizontalScrollIndicator={false}
        contentContainerStyle={styles.scrollContent}
      >
        {entities.map((entity) => (
          <Pressable
            key={entity.id}
            style={[styles.tag, { borderColor: getTypeColor(entity.type) + '60' }]}
            onPress={() => onEntityPress?.(entity.id)}
            onLongPress={() => setUnlinkTarget(entity)}
          >
            <Text style={[styles.tagType, { color: getTypeColor(entity.type) }]}>
              {entity.type}
            </Text>
            <Text style={styles.tagName}>{entity.name}</Text>
          </Pressable>
        ))}
        <Pressable style={styles.addTag} onPress={() => setShowSelector(true)}>
          <Text style={styles.addTagPlus}>+</Text>
          <Text style={styles.addTagText}>Tag</Text>
        </Pressable>
      </ScrollView>

      <EntitySelectorModal
        visible={showSelector}
        onDismiss={() => setShowSelector(false)}
        onSelect={handleLink}
        excludeIds={entities.map((e) => e.id)}
      />

      <ConfirmDialog
        visible={!!unlinkTarget}
        title="Lepas Entity"
        message={`Lepaskan "${unlinkTarget?.name}" dari dokumen ini?`}
        confirmText="Lepas"
        variant="danger"
        onConfirm={handleUnlink}
        onCancel={() => setUnlinkTarget(null)}
      />
    </View>
  );
}

// Entity selector modal - search and pick entity
function EntitySelectorModal({
  visible,
  onDismiss,
  onSelect,
  excludeIds,
}: {
  visible: boolean;
  onDismiss: () => void;
  onSelect: (entityId: string) => void;
  excludeIds: string[];
}) {
  const { searchEntities } = useEntityStore();
  const [query, setQuery] = useState('');
  const [results, setResults] = useState<Entity[]>([]);

  useEffect(() => {
    if (!visible) {
      setQuery('');
      setResults([]);
      return;
    }
    // Load all entities initially
    searchEntities('').then((all) => {
      // searchEntities returns empty for empty query, so load via list
    });
  }, [visible, searchEntities]);

  useEffect(() => {
    if (!visible) return;
    const timer = setTimeout(async () => {
      if (query.trim()) {
        const res = await searchEntities(query);
        setResults(res.filter((e) => !excludeIds.includes(e.id)));
      } else {
        setResults([]);
      }
    }, 300);
    return () => clearTimeout(timer);
  }, [query, visible, searchEntities, excludeIds]);

  return (
    <Modal visible={visible} transparent animationType="slide">
      <Pressable style={styles.selectorOverlay} onPress={onDismiss}>
        <Pressable style={styles.selectorSheet} onPress={() => {}}>
          <View style={styles.selectorHandle} />
          <Text style={styles.selectorTitle}>Pilih Entity</Text>
          <TextInput
            style={styles.selectorInput}
            value={query}
            onChangeText={setQuery}
            placeholder="Cari entity..."
            placeholderTextColor={colors.textMuted}
            autoFocus
          />
          <FlatList
            data={results}
            keyExtractor={(item) => item.id}
            renderItem={({ item }) => (
              <Pressable
                style={styles.selectorItem}
                onPress={() => onSelect(item.id)}
              >
                <View
                  style={[
                    styles.selectorBadge,
                    { backgroundColor: getTypeColor(item.type) },
                  ]}
                >
                  <Text style={styles.selectorBadgeText}>
                    {item.type.charAt(0)}
                  </Text>
                </View>
                <View>
                  <Text style={styles.selectorName}>{item.name}</Text>
                  <Text style={styles.selectorType}>{item.type}</Text>
                </View>
              </Pressable>
            )}
            ListEmptyComponent={
              query.trim() ? (
                <Text style={styles.selectorEmpty}>
                  Tidak ada entity yang cocok
                </Text>
              ) : (
                <Text style={styles.selectorEmpty}>
                  Ketik untuk mencari entity
                </Text>
              )
            }
            style={styles.selectorList}
          />
        </Pressable>
      </Pressable>
    </Modal>
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
    paddingVertical: spacing.xs,
  },
  scrollContent: {
    paddingHorizontal: spacing.lg,
    gap: spacing.sm,
    alignItems: 'center',
  },
  tag: {
    flexDirection: 'row',
    alignItems: 'center',
    paddingHorizontal: spacing.md,
    paddingVertical: spacing.xs + 2,
    borderRadius: 14,
    backgroundColor: colors.surface2,
    borderWidth: 1,
    gap: spacing.xs,
  },
  tagType: {
    fontSize: fontSize.xs,
    fontFamily: fontFamily.uiSemiBold,
  },
  tagName: {
    fontSize: fontSize.sm,
    fontFamily: fontFamily.ui,
    color: colors.textPrimary,
  },
  addTag: {
    flexDirection: 'row',
    alignItems: 'center',
    paddingHorizontal: spacing.md,
    paddingVertical: spacing.xs + 2,
    borderRadius: 14,
    borderWidth: 1,
    borderColor: colors.border,
    borderStyle: 'dashed',
    gap: spacing.xs,
  },
  addTagPlus: {
    fontSize: fontSize.sm,
    fontFamily: fontFamily.uiBold,
    color: colors.green,
  },
  addTagText: {
    fontSize: fontSize.sm,
    fontFamily: fontFamily.ui,
    color: colors.textMuted,
  },
  // Selector modal
  selectorOverlay: {
    flex: 1,
    backgroundColor: colors.overlay,
    justifyContent: 'flex-end',
  },
  selectorSheet: {
    backgroundColor: colors.surface,
    borderTopLeftRadius: 20,
    borderTopRightRadius: 20,
    maxHeight: '70%',
    paddingBottom: 40,
  },
  selectorHandle: {
    width: 40,
    height: 4,
    borderRadius: 2,
    backgroundColor: colors.textMuted,
    alignSelf: 'center',
    marginTop: spacing.md,
    marginBottom: spacing.lg,
  },
  selectorTitle: {
    fontSize: fontSize.lg,
    fontFamily: fontFamily.uiBold,
    color: colors.textPrimary,
    textAlign: 'center',
    marginBottom: spacing.md,
  },
  selectorInput: {
    fontFamily: fontFamily.ui,
    fontSize: fontSize.md,
    color: colors.textPrimary,
    backgroundColor: colors.surface2,
    paddingHorizontal: spacing.md,
    paddingVertical: spacing.md,
    borderRadius: 10,
    borderWidth: 1,
    borderColor: colors.border,
    marginHorizontal: spacing.lg,
    marginBottom: spacing.md,
  },
  selectorList: {
    paddingHorizontal: spacing.lg,
  },
  selectorItem: {
    flexDirection: 'row',
    alignItems: 'center',
    paddingVertical: spacing.md,
    borderBottomWidth: 1,
    borderBottomColor: colors.border,
    gap: spacing.md,
  },
  selectorBadge: {
    width: 36,
    height: 36,
    borderRadius: 18,
    justifyContent: 'center',
    alignItems: 'center',
  },
  selectorBadgeText: {
    fontSize: fontSize.sm,
    fontFamily: fontFamily.uiBold,
    color: colors.background,
  },
  selectorName: {
    fontSize: fontSize.md,
    fontFamily: fontFamily.uiSemiBold,
    color: colors.textPrimary,
  },
  selectorType: {
    fontSize: fontSize.xs,
    fontFamily: fontFamily.ui,
    color: colors.textMuted,
  },
  selectorEmpty: {
    fontSize: fontSize.sm,
    fontFamily: fontFamily.ui,
    color: colors.textMuted,
    textAlign: 'center',
    paddingVertical: spacing.xl,
  },
});
