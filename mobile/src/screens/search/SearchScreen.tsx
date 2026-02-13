// SearchScreen — global search with tabs
import React, { useState, useCallback } from 'react';
import {
  View,
  FlatList,
  ActivityIndicator,
  StyleSheet,
} from 'react-native';
import { useNavigation } from '@react-navigation/native';
import type { NativeStackNavigationProp } from '@react-navigation/native-stack';
import type { ChatStackParamList } from '@/navigation/types';
import { ScreenContainer } from '@/components/shared/ScreenContainer';
import { SearchBar } from '@/components/shared/SearchBar';
import { SearchTabBar } from '@/components/shared/SearchTabBar';
import { EmptyState } from '@/components/shared/EmptyState';
import {
  MessageResultItem,
  DocumentResultItem,
  ContactResultItem,
  EntityResultItem,
  SectionHeader,
} from '@/components/shared/SearchResultItems';
import { useSearch } from '@/hooks/useSearch';
import type { SearchTab } from '@/hooks/useSearch';
import type {
  MessageSearchResult,
  DocumentSearchResult,
  ContactSearchResult,
  EntitySearchResult,
} from '@/services/api/search';
import { colors, spacing } from '@/theme';

type NavigationProp = NativeStackNavigationProp<ChatStackParamList>;

export function SearchScreen() {
  const navigation = useNavigation<NavigationProp>();
  const [query, setQuery] = useState('');
  const [activeTab, setActiveTab] = useState<SearchTab>('all');

  const {
    allResults,
    messages,
    documents,
    contacts,
    entities,
    isLoading,
  } = useSearch(query, activeTab);

  const handleMessagePress = useCallback(
    (item: MessageSearchResult) => {
      navigation.navigate('Chat', {
        chatId: item.chatId,
        chatType: 'personal',
      });
    },
    [navigation],
  );

  const handleDocumentPress = useCallback(
    (item: DocumentSearchResult) => {
      navigation.navigate('DocumentEditor', {
        documentId: item.id,
      });
    },
    [navigation],
  );

  const handleContactPress = useCallback(
    (_item: ContactSearchResult) => {
      navigation.navigate('ContactList');
    },
    [navigation],
  );

  const handleEntityPress = useCallback(
    (_item: EntitySearchResult) => {
      // Navigate back — entity detail is in DocumentStack
    },
    [],
  );

  const handleSwitchTab = useCallback((tab: SearchTab) => {
    setActiveTab(tab);
  }, []);

  // Render content based on active tab
  const renderContent = () => {
    if (query.length < 2) {
      return (
        <EmptyState
          emoji="🔍"
          title="Cari di Chatat"
          description="Ketik minimal 2 karakter untuk mencari pesan, dokumen, kontak, atau entity."
        />
      );
    }

    if (isLoading) {
      return (
        <View style={styles.center}>
          <ActivityIndicator color={colors.green} size="large" />
        </View>
      );
    }

    switch (activeTab) {
      case 'all':
        return renderAllResults();
      case 'messages':
        return renderList(messages, renderMessageItem, 'pesan');
      case 'documents':
        return renderList(documents, renderDocItem, 'dokumen');
      case 'contacts':
        return renderList(contacts, renderContactItem, 'kontak');
      case 'entities':
        return renderList(entities, renderEntityItem, 'entity');
    }
  };

  const renderMessageItem = ({ item }: { item: MessageSearchResult }) => (
    <MessageResultItem item={item} onPress={handleMessagePress} />
  );

  const renderDocItem = ({ item }: { item: DocumentSearchResult }) => (
    <DocumentResultItem item={item} onPress={handleDocumentPress} />
  );

  const renderContactItem = ({ item }: { item: ContactSearchResult }) => (
    <ContactResultItem item={item} onPress={handleContactPress} />
  );

  const renderEntityItem = ({ item }: { item: EntitySearchResult }) => (
    <EntityResultItem item={item} onPress={handleEntityPress} />
  );

  function renderList<T extends { id: string }>(
    data: T[],
    renderItem: ({ item }: { item: T }) => React.ReactElement,
    type: string,
  ) {
    if (data.length === 0) {
      return (
        <EmptyState
          emoji="😕"
          title="Tidak ada hasil"
          description={`Tidak ditemukan ${type} untuk "${query}"`}
        />
      );
    }

    return (
      <FlatList
        data={data}
        keyExtractor={(item) => item.id}
        renderItem={renderItem}
        contentContainerStyle={styles.list}
      />
    );
  }

  const renderAllResults = () => {
    const { messages: msgs, documents: docs, contacts: ctcs, entities: ents } = allResults;
    const hasAny = msgs.length > 0 || docs.length > 0 || ctcs.length > 0 || ents.length > 0;

    if (!hasAny) {
      return (
        <EmptyState
          emoji="😕"
          title="Tidak ada hasil"
          description={`Tidak ditemukan hasil untuk "${query}"`}
        />
      );
    }

    type AllItem =
      | { type: 'msg_header' }
      | { type: 'msg'; data: MessageSearchResult }
      | { type: 'doc_header' }
      | { type: 'doc'; data: DocumentSearchResult }
      | { type: 'ctc_header' }
      | { type: 'ctc'; data: ContactSearchResult }
      | { type: 'ent_header' }
      | { type: 'ent'; data: EntitySearchResult };

    const items: AllItem[] = [];

    if (msgs.length > 0) {
      items.push({ type: 'msg_header' });
      msgs.forEach((m) => items.push({ type: 'msg', data: m }));
    }
    if (docs.length > 0) {
      items.push({ type: 'doc_header' });
      docs.forEach((d) => items.push({ type: 'doc', data: d }));
    }
    if (ctcs.length > 0) {
      items.push({ type: 'ctc_header' });
      ctcs.forEach((c) => items.push({ type: 'ctc', data: c }));
    }
    if (ents.length > 0) {
      items.push({ type: 'ent_header' });
      ents.forEach((e) => items.push({ type: 'ent', data: e }));
    }

    return (
      <FlatList
        data={items}
        keyExtractor={(item, index) => {
          if ('data' in item && 'id' in item.data) return item.data.id;
          return `header-${item.type}-${index}`;
        }}
        renderItem={({ item }) => {
          switch (item.type) {
            case 'msg_header':
              return (
                <SectionHeader
                  title="Pesan"
                  count={msgs.length}
                  onSeeAll={() => setActiveTab('messages')}
                />
              );
            case 'msg':
              return <MessageResultItem item={item.data} onPress={handleMessagePress} />;
            case 'doc_header':
              return (
                <SectionHeader
                  title="Dokumen"
                  count={docs.length}
                  onSeeAll={() => setActiveTab('documents')}
                />
              );
            case 'doc':
              return <DocumentResultItem item={item.data} onPress={handleDocumentPress} />;
            case 'ctc_header':
              return (
                <SectionHeader
                  title="Kontak"
                  count={ctcs.length}
                  onSeeAll={() => setActiveTab('contacts')}
                />
              );
            case 'ctc':
              return <ContactResultItem item={item.data} onPress={handleContactPress} />;
            case 'ent_header':
              return (
                <SectionHeader
                  title="Entity"
                  count={ents.length}
                  onSeeAll={() => setActiveTab('entities')}
                />
              );
            case 'ent':
              return <EntityResultItem item={item.data} onPress={handleEntityPress} />;
          }
        }}
        contentContainerStyle={styles.list}
      />
    );
  };

  return (
    <ScreenContainer>
      <View style={styles.header}>
        <SearchBar
          value={query}
          onChangeText={setQuery}
          placeholder="Cari pesan, dokumen, kontak..."
        />
      </View>
      <SearchTabBar activeTab={activeTab} onTabChange={handleSwitchTab} />
      <View style={styles.body}>{renderContent()}</View>
    </ScreenContainer>
  );
}

const styles = StyleSheet.create({
  header: {
    paddingHorizontal: spacing.lg,
  },
  body: {
    flex: 1,
  },
  center: {
    flex: 1,
    justifyContent: 'center',
    alignItems: 'center',
  },
  list: {
    paddingBottom: spacing.xxxl,
  },
});
