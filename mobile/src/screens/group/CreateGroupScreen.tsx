// Create group screen — 2-step wizard: select members, then enter group details
import React, { useCallback, useState, useMemo } from 'react';
import {
  View,
  Text,
  FlatList,
  Pressable,
  TextInput,
  ScrollView,
  Alert,
  ActivityIndicator,
  StyleSheet,
} from 'react-native';
import { SafeAreaView } from 'react-native-safe-area-context';
import type { NativeStackScreenProps } from '@react-navigation/native-stack';
import type { ChatStackParamList } from '@/navigation/types';
import { SearchBar } from '@/components/shared/SearchBar';
import { Avatar } from '@/components/ui/Avatar';
import { useContactStore } from '@/stores/contactStore';
import { useTranslation } from 'react-i18next';
import { chatsApi } from '@/services/api/chats';
import { colors, fontSize, fontFamily, spacing } from '@/theme';
import type { ContactInfo } from '@/types/chat';

type Props = NativeStackScreenProps<ChatStackParamList, 'CreateGroup'>;

const EMOJI_OPTIONS = ['💼', '🎯', '🚀', '💡', '🎮', '📚', '🏠', '❤️', '⭐', '🔥', '🎵', '🌟', '👥', '💬', '🛠️', '🎨'];

export function CreateGroupScreen({ navigation }: Props) {
  const { t } = useTranslation();
  const { contacts, fetchContacts } = useContactStore();

  // Step state: 1 = select members, 2 = group details
  const [step, setStep] = useState<1 | 2>(1);
  const [searchQuery, setSearchQuery] = useState('');
  const [selectedMembers, setSelectedMembers] = useState<ContactInfo[]>([]);
  const [groupName, setGroupName] = useState('');
  const [groupIcon, setGroupIcon] = useState('💼');
  const [groupDescription, setGroupDescription] = useState('');
  const [isCreating, setIsCreating] = useState(false);

  // Fetch contacts on mount
  React.useEffect(() => {
    fetchContacts();
  }, [fetchContacts]);

  // Filter contacts by search
  const filteredContacts = useMemo(() => {
    if (!searchQuery.trim()) return contacts;
    const query = searchQuery.toLowerCase();
    return contacts.filter(
      (c) =>
        (c.name ?? '').toLowerCase().includes(query) ||
        c.phone.includes(query),
    );
  }, [contacts, searchQuery]);

  const toggleMember = useCallback((contact: ContactInfo) => {
    setSelectedMembers((prev) => {
      const exists = prev.find((c) => c.userId === contact.userId);
      if (exists) return prev.filter((c) => c.userId !== contact.userId);
      return [...prev, contact];
    });
  }, []);

  const removeMember = useCallback((userId: string) => {
    setSelectedMembers((prev) => prev.filter((c) => c.userId !== userId));
  }, []);

  const handleNext = useCallback(() => {
    if (selectedMembers.length < 2) {
      Alert.alert(t('group.addMembers'), t('group.memberCount', { count: 2 }));
      return;
    }
    setStep(2);
  }, [selectedMembers]);

  const handleCreate = useCallback(async () => {
    if (!groupName.trim()) {
      Alert.alert(t('group.groupName'), t('group.groupName'));
      return;
    }

    setIsCreating(true);
    try {
      const res = await chatsApi.createGroup({
        type: 'group',
        name: groupName.trim(),
        icon: groupIcon,
        description: groupDescription.trim(),
        memberIds: selectedMembers.map((m) => m.userId),
      });
      const chat = res.data.data;
      navigation.replace('Chat', { chatId: chat.id, chatType: 'group' });
    } catch {
      Alert.alert(t('common.error'), t('common.retry'));
    } finally {
      setIsCreating(false);
    }
  }, [groupName, groupIcon, groupDescription, selectedMembers, navigation]);

  // Update nav header
  React.useEffect(() => {
    navigation.setOptions({
      title: step === 1 ? t('group.addMembers') : t('group.groupInfo'),
      headerLeft: step === 2
        ? () => (
            <Pressable onPress={() => setStep(1)} style={{ marginRight: spacing.md }}>
              <Text style={{ color: colors.green, fontFamily: fontFamily.ui, fontSize: fontSize.md }}>
                {t('common.back')}
              </Text>
            </Pressable>
          )
        : undefined,
    });
  }, [navigation, step]);

  if (step === 1) {
    return (
      <SafeAreaView style={styles.container} edges={['bottom']}>
        {/* Selected member chips */}
        {selectedMembers.length > 0 && (
          <ScrollView
            horizontal
            showsHorizontalScrollIndicator={false}
            style={styles.chipContainer}
            contentContainerStyle={styles.chipContent}
          >
            {selectedMembers.map((member) => (
              <Pressable
                key={member.userId}
                style={styles.chip}
                onPress={() => removeMember(member.userId)}
              >
                <Avatar emoji={member.avatar || '\u{1F464}'} size="sm" />
                <Text style={styles.chipName} numberOfLines={1}>
                  {member.name ?? member.phone}
                </Text>
                <Text style={styles.chipRemove}>✕</Text>
              </Pressable>
            ))}
          </ScrollView>
        )}

        <SearchBar
          value={searchQuery}
          onChangeText={setSearchQuery}
          placeholder={t('contact.searchContacts')}
        />

        <FlatList
          data={filteredContacts}
          keyExtractor={(item) => item.userId}
          renderItem={({ item }) => {
            const isSelected = selectedMembers.some((m) => m.userId === item.userId);
            return (
              <Pressable
                style={[styles.contactRow, isSelected && styles.contactRowSelected]}
                onPress={() => toggleMember(item)}
              >
                <Avatar
                  emoji={item.avatar || '\u{1F464}'}
                  size="md"
                  online={item.isOnline}
                />
                <View style={styles.contactInfo}>
                  <Text style={styles.contactName}>{item.name ?? item.phone}</Text>
                  <Text style={styles.contactStatus} numberOfLines={1}>
                    {item.status || item.phone}
                  </Text>
                </View>
                <View style={[styles.checkbox, isSelected && styles.checkboxSelected]}>
                  {isSelected && <Text style={styles.checkmark}>✓</Text>}
                </View>
              </Pressable>
            );
          }}
          contentContainerStyle={styles.listContent}
        />

        <Pressable
          style={[
            styles.nextButton,
            selectedMembers.length < 2 && styles.nextButtonDisabled,
          ]}
          onPress={handleNext}
          disabled={selectedMembers.length < 2}
        >
          <Text style={styles.nextButtonText}>
            {t('common.next')} ({selectedMembers.length})
          </Text>
        </Pressable>
      </SafeAreaView>
    );
  }

  // Step 2: Group Details
  return (
    <SafeAreaView style={styles.container} edges={['bottom']}>
      <ScrollView style={styles.detailsContainer} contentContainerStyle={styles.detailsContent}>
        {/* Emoji picker */}
        <Text style={styles.sectionLabel}>{t('group.groupName')}</Text>
        <View style={styles.emojiGrid}>
          {EMOJI_OPTIONS.map((emoji) => (
            <Pressable
              key={emoji}
              style={[
                styles.emojiOption,
                groupIcon === emoji && styles.emojiOptionSelected,
              ]}
              onPress={() => setGroupIcon(emoji)}
            >
              <Text style={styles.emojiText}>{emoji}</Text>
            </Pressable>
          ))}
        </View>

        {/* Group name */}
        <Text style={styles.sectionLabel}>{t('group.groupName')}</Text>
        <TextInput
          style={styles.textInput}
          value={groupName}
          onChangeText={setGroupName}
          placeholder={t('group.groupName')}
          placeholderTextColor={colors.textMuted}
          maxLength={100}
        />

        {/* Description */}
        <Text style={styles.sectionLabel}>{t('common.edit')}</Text>
        <TextInput
          style={[styles.textInput, styles.textArea]}
          value={groupDescription}
          onChangeText={setGroupDescription}
          placeholder={t('common.edit')}
          placeholderTextColor={colors.textMuted}
          multiline
          numberOfLines={3}
        />

        {/* Selected members preview */}
        <Text style={styles.sectionLabel}>{t('group.members')} ({selectedMembers.length + 1})</Text>
        <View style={styles.memberPreview}>
          {selectedMembers.map((member) => (
            <View key={member.userId} style={styles.memberPreviewItem}>
              <Avatar emoji={member.avatar || '\u{1F464}'} size="sm" />
              <Text style={styles.memberPreviewName} numberOfLines={1}>
                {member.name ?? member.phone}
              </Text>
            </View>
          ))}
        </View>
      </ScrollView>

      <Pressable
        style={[styles.createButton, isCreating && styles.createButtonDisabled]}
        onPress={handleCreate}
        disabled={isCreating || !groupName.trim()}
      >
        {isCreating ? (
          <ActivityIndicator color={colors.background} size="small" />
        ) : (
          <Text style={styles.createButtonText}>{t('group.createGroup')}</Text>
        )}
      </Pressable>
    </SafeAreaView>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: colors.background,
  },
  chipContainer: {
    maxHeight: 72,
    paddingVertical: spacing.xs,
  },
  chipContent: {
    paddingHorizontal: spacing.md,
    gap: spacing.xs,
  },
  chip: {
    flexDirection: 'row',
    alignItems: 'center',
    backgroundColor: colors.surface2,
    borderRadius: 20,
    paddingHorizontal: spacing.sm,
    paddingVertical: spacing.xs,
    marginRight: spacing.xs,
    gap: spacing.xs,
  },
  chipName: {
    fontFamily: fontFamily.ui,
    fontSize: fontSize.sm,
    color: colors.textPrimary,
    maxWidth: 80,
  },
  chipRemove: {
    fontFamily: fontFamily.ui,
    fontSize: fontSize.sm,
    color: colors.textMuted,
  },
  contactRow: {
    flexDirection: 'row',
    alignItems: 'center',
    paddingHorizontal: spacing.md,
    paddingVertical: spacing.sm,
    gap: spacing.md,
  },
  contactRowSelected: {
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
  checkbox: {
    width: 24,
    height: 24,
    borderRadius: 12,
    borderWidth: 2,
    borderColor: colors.border,
    justifyContent: 'center',
    alignItems: 'center',
  },
  checkboxSelected: {
    backgroundColor: colors.green,
    borderColor: colors.green,
  },
  checkmark: {
    color: colors.background,
    fontSize: 14,
    fontWeight: 'bold',
  },
  listContent: {
    paddingBottom: spacing.xl,
  },
  nextButton: {
    backgroundColor: colors.green,
    marginHorizontal: spacing.md,
    marginBottom: spacing.md,
    paddingVertical: spacing.md,
    borderRadius: 12,
    alignItems: 'center',
  },
  nextButtonDisabled: {
    opacity: 0.5,
  },
  nextButtonText: {
    fontFamily: fontFamily.uiSemiBold,
    fontSize: fontSize.md,
    color: colors.background,
  },
  detailsContainer: {
    flex: 1,
  },
  detailsContent: {
    padding: spacing.md,
  },
  sectionLabel: {
    fontFamily: fontFamily.uiSemiBold,
    fontSize: fontSize.sm,
    color: colors.textMuted,
    marginTop: spacing.lg,
    marginBottom: spacing.sm,
  },
  emojiGrid: {
    flexDirection: 'row',
    flexWrap: 'wrap',
    gap: spacing.sm,
  },
  emojiOption: {
    width: 48,
    height: 48,
    borderRadius: 12,
    backgroundColor: colors.surface,
    justifyContent: 'center',
    alignItems: 'center',
  },
  emojiOptionSelected: {
    backgroundColor: colors.green,
  },
  emojiText: {
    fontSize: 24,
  },
  textInput: {
    backgroundColor: colors.surface,
    borderRadius: 12,
    paddingHorizontal: spacing.md,
    paddingVertical: spacing.sm,
    fontFamily: fontFamily.ui,
    fontSize: fontSize.md,
    color: colors.textPrimary,
  },
  textArea: {
    minHeight: 80,
    textAlignVertical: 'top',
  },
  memberPreview: {
    flexDirection: 'row',
    flexWrap: 'wrap',
    gap: spacing.sm,
  },
  memberPreviewItem: {
    alignItems: 'center',
    width: 60,
  },
  memberPreviewName: {
    fontFamily: fontFamily.ui,
    fontSize: fontSize.xs,
    color: colors.textMuted,
    marginTop: 4,
    textAlign: 'center',
  },
  createButton: {
    backgroundColor: colors.green,
    marginHorizontal: spacing.md,
    marginBottom: spacing.md,
    paddingVertical: spacing.md,
    borderRadius: 12,
    alignItems: 'center',
  },
  createButtonDisabled: {
    opacity: 0.5,
  },
  createButtonText: {
    fontFamily: fontFamily.uiSemiBold,
    fontSize: fontSize.md,
    color: colors.background,
  },
});
