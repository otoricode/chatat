// Documents tab content — shows documents belonging to a chat
import React, { useCallback, useEffect, useState } from 'react';
import {
  View,
  FlatList,
  ActivityIndicator,
  TouchableOpacity,
  Text,
  StyleSheet,
} from 'react-native';
import { useNavigation } from '@react-navigation/native';
import type { NativeStackNavigationProp } from '@react-navigation/native-stack';
import type { ChatStackParamList } from '@/navigation/types';
import { DocumentCard } from '@/components/document/DocumentCard';
import { EmptyState } from '@/components/shared/EmptyState';
import { documentsApi } from '@/services/api/documents';
import { useTranslation } from 'react-i18next';
import { formatMessageTime } from '@/lib/timeFormat';
import { colors, fontSize, fontFamily, spacing } from '@/theme';
import type { DocumentListItem } from '@/types/chat';

type NavigationProp = NativeStackNavigationProp<ChatStackParamList>;

interface ChatDocumentsTabProps {
  chatId: string;
}

export function ChatDocumentsTab({ chatId }: ChatDocumentsTabProps) {
  const { t } = useTranslation();
  const navigation = useNavigation<NavigationProp>();
  const [documents, setDocuments] = useState<DocumentListItem[]>([]);
  const [isLoading, setIsLoading] = useState(true);

  const fetchDocuments = useCallback(async () => {
    try {
      setIsLoading(true);
      const res = await documentsApi.listByChat(chatId);
      setDocuments(res.data.data ?? []);
    } catch {
      // Silent fail — empty state will show
    } finally {
      setIsLoading(false);
    }
  }, [chatId]);

  useEffect(() => {
    fetchDocuments();
  }, [fetchDocuments]);

  const handleDocumentPress = useCallback(
    (doc: DocumentListItem) => {
      navigation.navigate('DocumentEditor', { documentId: doc.id });
    },
    [navigation],
  );

  const handleCreate = useCallback(() => {
    navigation.navigate('DocumentEditor', {
      contextType: 'chat',
      contextId: chatId,
    });
  }, [navigation, chatId]);

  const renderItem = useCallback(
    ({ item }: { item: DocumentListItem }) => (
      <View style={styles.cardWrapper}>
        <DocumentCard
          title={item.title}
          icon={item.icon}
          locked={item.locked}
          updatedAt={formatMessageTime(item.updatedAt)}
          onPress={() => handleDocumentPress(item)}
        />
      </View>
    ),
    [handleDocumentPress],
  );

  const keyExtractor = useCallback((item: DocumentListItem) => item.id, []);

  if (isLoading) {
    return (
      <View style={styles.center}>
        <ActivityIndicator color={colors.green} />
      </View>
    );
  }

  return (
    <View style={styles.container}>
      {documents.length === 0 ? (
        <View style={styles.emptyContainer}>
          <EmptyState
            emoji="\u{1F4C4}"
            title={t('document.noDocuments')}
            description={t('document.createFirst')}
          />
          <TouchableOpacity style={styles.createButton} onPress={handleCreate}>
            <Text style={styles.createButtonText}>{t('document.newDocument')}</Text>
          </TouchableOpacity>
        </View>
      ) : (
        <>
          <FlatList
            data={documents}
            renderItem={renderItem}
            keyExtractor={keyExtractor}
            contentContainerStyle={styles.listContent}
          />
          <TouchableOpacity style={styles.fab} onPress={handleCreate}>
            <Text style={styles.fabIcon}>+</Text>
          </TouchableOpacity>
        </>
      )}
    </View>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: colors.background,
  },
  center: {
    flex: 1,
    justifyContent: 'center',
    alignItems: 'center',
  },
  emptyContainer: {
    flex: 1,
    justifyContent: 'center',
    alignItems: 'center',
    padding: spacing.xl,
  },
  listContent: {
    padding: spacing.md,
    gap: spacing.sm,
  },
  cardWrapper: {
    marginBottom: spacing.sm,
  },
  createButton: {
    marginTop: spacing.lg,
    backgroundColor: colors.green,
    paddingHorizontal: spacing.xl,
    paddingVertical: spacing.md,
    borderRadius: 12,
  },
  createButtonText: {
    fontFamily: fontFamily.uiSemiBold,
    fontSize: fontSize.md,
    color: colors.white,
  },
  fab: {
    position: 'absolute',
    bottom: spacing.lg,
    right: spacing.lg,
    width: 56,
    height: 56,
    borderRadius: 28,
    backgroundColor: colors.green,
    justifyContent: 'center',
    alignItems: 'center',
    elevation: 4,
    shadowColor: '#000',
    shadowOffset: { width: 0, height: 2 },
    shadowOpacity: 0.25,
    shadowRadius: 4,
  },
  fabIcon: {
    fontSize: 28,
    color: colors.white,
    lineHeight: 30,
  },
});
