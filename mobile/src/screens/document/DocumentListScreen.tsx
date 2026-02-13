// Document list screen — main document tab
import React, { useEffect, useState, useCallback } from 'react';
import {
  View,
  Text,
  FlatList,
  Pressable,
  RefreshControl,
  StyleSheet,
} from 'react-native';
import { SafeAreaView } from 'react-native-safe-area-context';
import type { NativeStackScreenProps } from '@react-navigation/native-stack';
import type { DocumentStackParamList } from '@/navigation/types';
import type { DocumentListItem } from '@/types/chat';
import { Header } from '@/components/shared/Header';
import { FAB } from '@/components/shared/FAB';
import { EmptyState } from '@/components/shared/EmptyState';
import { LockStatusBadge } from '@/components/document/LockStatusBadge';
import { documentsApi } from '@/services/api/documents';
import { colors, fontFamily, fontSize, spacing } from '@/theme';
import { formatDistanceToNow } from 'date-fns';
import { id as idLocale } from 'date-fns/locale';

type Props = NativeStackScreenProps<DocumentStackParamList, 'DocumentList'>;

export function DocumentListScreen({ navigation }: Props) {
  const [documents, setDocuments] = useState<DocumentListItem[]>([]);
  const [isLoading, setIsLoading] = useState(false);

  const fetchDocuments = useCallback(async () => {
    setIsLoading(true);
    try {
      const res = await documentsApi.list();
      setDocuments(res.data.data ?? []);
    } catch {
      // Silent fail
    } finally {
      setIsLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchDocuments();
  }, [fetchDocuments]);

  // Refresh on focus
  useEffect(() => {
    const unsubscribe = navigation.addListener('focus', () => {
      fetchDocuments();
    });
    return unsubscribe;
  }, [navigation, fetchDocuments]);

  const handleNewDocument = () => {
    navigation.navigate('DocumentEditor', {});
  };

  const handleOpenDocument = (item: DocumentListItem) => {
    if (item.locked) {
      navigation.navigate('DocumentViewer', { documentId: item.id });
    } else {
      navigation.navigate('DocumentEditor', { documentId: item.id });
    }
  };

  const renderItem = ({ item }: { item: DocumentListItem }) => {
    const timeAgo = formatDistanceToNow(new Date(item.updatedAt), {
      addSuffix: true,
      locale: idLocale,
    });

    return (
      <Pressable
        style={({ pressed }) => [styles.docItem, pressed && styles.docItemPressed]}
        onPress={() => handleOpenDocument(item)}
      >
        <Text style={styles.docIcon}>{item.icon || '📄'}</Text>
        <View style={styles.docInfo}>
          <Text style={styles.docTitle} numberOfLines={1}>
            {item.title || 'Tanpa Judul'}
          </Text>
          <Text style={styles.docMeta}>{timeAgo}</Text>
        </View>
        {item.locked && <LockStatusBadge locked={item.locked} compact />}
      </Pressable>
    );
  };

  return (
    <SafeAreaView style={styles.container} edges={['top']}>
      <Header title="Dokumen" />
      <View style={styles.content}>
        {documents.length === 0 && !isLoading ? (
          <EmptyState
            emoji="📄"
            title="Belum ada dokumen"
            description="Buat dokumen baru untuk mulai berkolaborasi"
          />
        ) : (
          <FlatList
            data={documents}
            keyExtractor={(item) => item.id}
            renderItem={renderItem}
            refreshControl={
              <RefreshControl
                refreshing={isLoading}
                onRefresh={fetchDocuments}
                tintColor={colors.green}
              />
            }
            contentContainerStyle={styles.listContent}
          />
        )}
      </View>
      <FAB onPress={handleNewDocument} />
    </SafeAreaView>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: colors.background,
  },
  content: {
    flex: 1,
  },
  listContent: {
    paddingHorizontal: spacing.lg,
    paddingVertical: spacing.sm,
  },
  docItem: {
    flexDirection: 'row',
    alignItems: 'center',
    paddingVertical: spacing.md,
    paddingHorizontal: spacing.md,
    borderRadius: 10,
    marginBottom: spacing.xs,
  },
  docItemPressed: {
    backgroundColor: colors.surface,
  },
  docIcon: {
    fontSize: 24,
    marginRight: spacing.md,
  },
  docInfo: {
    flex: 1,
  },
  docTitle: {
    fontFamily: fontFamily.uiMedium,
    fontSize: fontSize.md,
    color: colors.textPrimary,
  },
  docMeta: {
    fontFamily: fontFamily.ui,
    fontSize: fontSize.xs,
    color: colors.textMuted,
    marginTop: 2,
  },
});
