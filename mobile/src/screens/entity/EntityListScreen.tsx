// Entity list screen — view and manage entities
import React, { useCallback, useEffect, useMemo, useState } from 'react';
import {
  View,
  Text,
  FlatList,
  Pressable,
  ScrollView,
  StyleSheet,
  RefreshControl,
} from 'react-native';
import { SafeAreaView } from 'react-native-safe-area-context';
import type { NativeStackScreenProps } from '@react-navigation/native-stack';
import type { DocumentStackParamList } from '@/navigation/types';
import { SearchBar } from '@/components/shared/SearchBar';
import { EmptyState } from '@/components/shared/EmptyState';
import { FAB } from '@/components/shared/FAB';
import { useEntityStore } from '@/stores/entityStore';
import { colors, fontSize, fontFamily, spacing } from '@/theme';
import type { EntityListItem as EntityListItemType } from '@/types/chat';
import { CreateEntitySheet } from '@/components/entity/CreateEntitySheet';

type Props = NativeStackScreenProps<DocumentStackParamList, 'EntityList'>;

export function EntityListScreen({ navigation }: Props) {
  const { entities, types, isLoading, fetchEntities, fetchTypes } = useEntityStore();
  const [searchQuery, setSearchQuery] = useState('');
  const [typeFilter, setTypeFilter] = useState<string | null>(null);
  const [refreshing, setRefreshing] = useState(false);
  const [showCreate, setShowCreate] = useState(false);

  useEffect(() => {
    fetchEntities();
    fetchTypes();
  }, [fetchEntities, fetchTypes]);

  const handleRefresh = useCallback(async () => {
    setRefreshing(true);
    await fetchEntities(typeFilter ?? undefined);
    await fetchTypes();
    setRefreshing(false);
  }, [fetchEntities, fetchTypes, typeFilter]);

  const handleTypeFilter = useCallback(
    (type: string | null) => {
      setTypeFilter(type);
      fetchEntities(type ?? undefined);
    },
    [fetchEntities],
  );

  const filteredEntities = useMemo(() => {
    if (!searchQuery.trim()) return entities;
    const lower = searchQuery.toLowerCase();
    return entities.filter(
      (e) =>
        e.name.toLowerCase().includes(lower) ||
        e.type.toLowerCase().includes(lower),
    );
  }, [entities, searchQuery]);

  const handleEntityPress = useCallback(
    (entity: EntityListItemType) => {
      navigation.navigate('EntityDetail', { entityId: entity.id });
    },
    [navigation],
  );

  const handleCreated = useCallback(() => {
    setShowCreate(false);
    fetchEntities(typeFilter ?? undefined);
    fetchTypes();
  }, [fetchEntities, fetchTypes, typeFilter]);

  const renderItem = useCallback(
    ({ item }: { item: EntityListItemType }) => (
      <Pressable
        style={({ pressed }) => [styles.item, pressed && styles.itemPressed]}
        onPress={() => handleEntityPress(item)}
      >
        <View style={styles.itemLeft}>
          <View style={[styles.typeBadge, { backgroundColor: getTypeColor(item.type) }]}>
            <Text style={styles.typeBadgeText}>{item.type.charAt(0)}</Text>
          </View>
          <View style={styles.itemInfo}>
            <Text style={styles.itemName} numberOfLines={1}>
              {item.name}
            </Text>
            <Text style={styles.itemType}>{item.type}</Text>
          </View>
        </View>
        {item.documentCount > 0 && (
          <View style={styles.docCount}>
            <Text style={styles.docCountText}>{item.documentCount} dok</Text>
          </View>
        )}
      </Pressable>
    ),
    [handleEntityPress],
  );

  return (
    <SafeAreaView style={styles.container} edges={['top']}>
      <View style={styles.header}>
        <Text style={styles.title}>Entity</Text>
      </View>

      <SearchBar
        placeholder="Cari entity..."
        value={searchQuery}
        onChangeText={setSearchQuery}
      />

      <ScrollView
        horizontal
        showsHorizontalScrollIndicator={false}
        style={styles.filterRow}
        contentContainerStyle={styles.filterContent}
      >
        <Pressable
          style={[styles.chip, !typeFilter && styles.chipActive]}
          onPress={() => handleTypeFilter(null)}
        >
          <Text style={[styles.chipText, !typeFilter && styles.chipTextActive]}>
            Semua
          </Text>
        </Pressable>
        {types.map((t) => (
          <Pressable
            key={t}
            style={[styles.chip, typeFilter === t && styles.chipActive]}
            onPress={() => handleTypeFilter(t)}
          >
            <Text
              style={[styles.chipText, typeFilter === t && styles.chipTextActive]}
            >
              {t}
            </Text>
          </Pressable>
        ))}
      </ScrollView>

      <FlatList
        data={filteredEntities}
        keyExtractor={(item) => item.id}
        renderItem={renderItem}
        contentContainerStyle={styles.list}
        refreshControl={
          <RefreshControl
            refreshing={refreshing}
            onRefresh={handleRefresh}
            tintColor={colors.green}
          />
        }
        ListEmptyComponent={
          !isLoading ? (
            <EmptyState
              emoji="🏷"
              title="Belum ada entity"
              description="Buat entity untuk menandai dan mengelola data di dokumen."
            />
          ) : null
        }
      />

      <FAB onPress={() => setShowCreate(true)} />

      <CreateEntitySheet
        visible={showCreate}
        onDismiss={() => setShowCreate(false)}
        onCreated={handleCreated}
      />
    </SafeAreaView>
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
  header: {
    paddingHorizontal: spacing.lg,
    paddingVertical: spacing.md,
  },
  title: {
    fontSize: fontSize.xl,
    fontFamily: fontFamily.uiBold,
    color: colors.textPrimary,
  },
  filterRow: {
    maxHeight: 44,
  },
  filterContent: {
    paddingHorizontal: spacing.lg,
    gap: spacing.sm,
    alignItems: 'center',
  },
  chip: {
    paddingHorizontal: spacing.md,
    paddingVertical: spacing.xs + 2,
    borderRadius: 16,
    backgroundColor: colors.surface2,
    borderWidth: 1,
    borderColor: colors.border,
  },
  chipActive: {
    backgroundColor: colors.green + '20',
    borderColor: colors.green,
  },
  chipText: {
    fontSize: fontSize.sm,
    fontFamily: fontFamily.ui,
    color: colors.textMuted,
  },
  chipTextActive: {
    color: colors.green,
  },
  list: {
    paddingHorizontal: spacing.lg,
    paddingBottom: 80,
  },
  item: {
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'space-between',
    paddingVertical: spacing.md,
    borderBottomWidth: 1,
    borderBottomColor: colors.border,
  },
  itemPressed: {
    opacity: 0.7,
  },
  itemLeft: {
    flexDirection: 'row',
    alignItems: 'center',
    flex: 1,
  },
  typeBadge: {
    width: 40,
    height: 40,
    borderRadius: 20,
    justifyContent: 'center',
    alignItems: 'center',
    marginRight: spacing.md,
  },
  typeBadgeText: {
    fontSize: fontSize.md,
    fontFamily: fontFamily.uiBold,
    color: colors.background,
  },
  itemInfo: {
    flex: 1,
  },
  itemName: {
    fontSize: fontSize.md,
    fontFamily: fontFamily.uiSemiBold,
    color: colors.textPrimary,
  },
  itemType: {
    fontSize: fontSize.sm,
    fontFamily: fontFamily.ui,
    color: colors.textMuted,
    marginTop: 2,
  },
  docCount: {
    paddingHorizontal: spacing.sm,
    paddingVertical: spacing.xs,
    borderRadius: 8,
    backgroundColor: colors.surface2,
  },
  docCountText: {
    fontSize: fontSize.xs,
    fontFamily: fontFamily.ui,
    color: colors.textMuted,
  },
});
