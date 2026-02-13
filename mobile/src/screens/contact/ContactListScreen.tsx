// Contact list screen — select contact to start chat
import React, { useCallback, useEffect, useState, useMemo } from 'react';
import {
  View,
  Text,
  FlatList,
  Pressable,
  StyleSheet,
  RefreshControl,
} from 'react-native';
import { SafeAreaView } from 'react-native-safe-area-context';
import type { NativeStackScreenProps } from '@react-navigation/native-stack';
import type { ChatStackParamList } from '@/navigation/types';
import { SearchBar } from '@/components/shared/SearchBar';
import { Avatar } from '@/components/ui/Avatar';
import { EmptyState } from '@/components/shared/EmptyState';
import { useContactStore } from '@/stores/contactStore';
import { chatsApi } from '@/services/api/chats';
import { colors, fontSize, fontFamily, spacing } from '@/theme';
import type { ContactInfo } from '@/types/chat';

type Props = NativeStackScreenProps<ChatStackParamList, 'ContactList'>;

type Section = { type: 'header'; letter: string } | { type: 'contact'; contact: ContactInfo };

export function ContactListScreen({ navigation }: Props) {
  const { contacts, isLoading, fetchContacts, searchContacts } = useContactStore();
  const [searchQuery, setSearchQuery] = useState('');
  const [refreshing, setRefreshing] = useState(false);

  useEffect(() => {
    fetchContacts();
  }, [fetchContacts]);

  const handleRefresh = useCallback(async () => {
    setRefreshing(true);
    await fetchContacts();
    setRefreshing(false);
  }, [fetchContacts]);

  const filteredContacts = useMemo(() => {
    return searchQuery ? searchContacts(searchQuery) : contacts;
  }, [contacts, searchQuery, searchContacts]);

  // Build sections with alphabetical headers
  const sections: Section[] = useMemo(() => {
    const result: Section[] = [];
    let currentLetter = '';

    for (const contact of filteredContacts) {
      const firstLetter = (contact.name ?? '?').charAt(0).toUpperCase();
      if (firstLetter !== currentLetter) {
        currentLetter = firstLetter;
        result.push({ type: 'header', letter: currentLetter });
      }
      result.push({ type: 'contact', contact });
    }

    return result;
  }, [filteredContacts]);

  const handleContactPress = useCallback(
    async (contact: ContactInfo) => {
      try {
        const res = await chatsApi.create(contact.userId);
        navigation.replace('Chat', {
          chatId: res.data.data.id,
          chatType: 'personal',
        });
      } catch {
        // If chat already exists, create returns it
        navigation.goBack();
      }
    },
    [navigation],
  );

  const handleCreateGroup = useCallback(() => {
    navigation.navigate('CreateGroup');
  }, [navigation]);

  const renderItem = useCallback(
    ({ item }: { item: Section }) => {
      if (item.type === 'header') {
        return (
          <View style={styles.sectionHeader}>
            <Text style={styles.sectionLetter}>{item.letter}</Text>
          </View>
        );
      }

      const { contact } = item;
      return (
        <Pressable
          style={({ pressed }) => [styles.contactItem, pressed && styles.pressed]}
          onPress={() => handleContactPress(contact)}
        >
          <Avatar
            emoji={contact.avatar || '\u{1F464}'}
            size="md"
            online={contact.isOnline}
          />
          <View style={styles.contactInfo}>
            <Text style={styles.contactName}>{contact.name}</Text>
            {contact.status ? (
              <Text style={styles.contactStatus} numberOfLines={1}>
                {contact.status}
              </Text>
            ) : null}
          </View>
        </Pressable>
      );
    },
    [handleContactPress],
  );

  const keyExtractor = useCallback(
    (item: Section, index: number) =>
      item.type === 'header' ? `hdr-${item.letter}` : `ct-${item.contact.userId}-${index}`,
    [],
  );

  return (
    <SafeAreaView style={styles.container} edges={['top']}>
      <View style={styles.headerRow}>
        <Pressable onPress={() => navigation.goBack()} style={styles.backButton}>
          <Text style={styles.backIcon}>{'\u2190'}</Text>
        </Pressable>
        <Text style={styles.title}>Pilih Kontak</Text>
      </View>

      <SearchBar
        value={searchQuery}
        onChangeText={setSearchQuery}
        placeholder="Cari kontak..."
      />

      <Pressable
        style={({ pressed }) => [styles.createGroupButton, pressed && styles.pressed]}
        onPress={handleCreateGroup}
      >
        <View style={styles.groupIconContainer}>
          <Text style={styles.groupIcon}>{'\u{1F465}'}</Text>
        </View>
        <Text style={styles.createGroupText}>Buat Grup Baru</Text>
      </Pressable>

      {filteredContacts.length === 0 && !isLoading ? (
        <EmptyState
          emoji="\u{1F465}"
          title="Belum ada kontak"
          description="Kontak yang menggunakan Chatat akan muncul di sini"
        />
      ) : (
        <FlatList
          data={sections}
          renderItem={renderItem}
          keyExtractor={keyExtractor}
          refreshControl={
            <RefreshControl
              refreshing={refreshing}
              onRefresh={handleRefresh}
              tintColor={colors.green}
              colors={[colors.green]}
            />
          }
        />
      )}
    </SafeAreaView>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: colors.background,
  },
  headerRow: {
    flexDirection: 'row',
    alignItems: 'center',
    paddingHorizontal: spacing.lg,
    paddingVertical: spacing.md,
    backgroundColor: colors.headerBackground,
  },
  backButton: {
    marginRight: spacing.md,
    padding: spacing.xs,
  },
  backIcon: {
    fontSize: 20,
    color: colors.textPrimary,
  },
  title: {
    fontFamily: fontFamily.uiBold,
    fontSize: fontSize.lg,
    color: colors.textPrimary,
  },
  createGroupButton: {
    flexDirection: 'row',
    alignItems: 'center',
    paddingHorizontal: spacing.lg,
    paddingVertical: spacing.md,
    gap: spacing.md,
  },
  groupIconContainer: {
    width: 44,
    height: 44,
    borderRadius: 22,
    backgroundColor: colors.green,
    justifyContent: 'center',
    alignItems: 'center',
  },
  groupIcon: {
    fontSize: 20,
  },
  createGroupText: {
    fontFamily: fontFamily.uiSemiBold,
    fontSize: fontSize.md,
    color: colors.textPrimary,
  },
  sectionHeader: {
    paddingHorizontal: spacing.lg,
    paddingVertical: spacing.xs,
    backgroundColor: colors.surface,
  },
  sectionLetter: {
    fontFamily: fontFamily.uiSemiBold,
    fontSize: fontSize.sm,
    color: colors.green,
  },
  contactItem: {
    flexDirection: 'row',
    alignItems: 'center',
    paddingHorizontal: spacing.lg,
    paddingVertical: spacing.md,
    gap: spacing.md,
  },
  pressed: {
    backgroundColor: colors.surface,
  },
  contactInfo: {
    flex: 1,
  },
  contactName: {
    fontFamily: fontFamily.uiSemiBold,
    fontSize: fontSize.md,
    color: colors.textPrimary,
  },
  contactStatus: {
    fontFamily: fontFamily.ui,
    fontSize: fontSize.sm,
    color: colors.textMuted,
    marginTop: 2,
  },
});
