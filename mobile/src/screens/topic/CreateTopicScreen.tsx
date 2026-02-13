// Create Topic Screen — wizard for creating a topic from a chat/group
import React, { useState, useEffect, useCallback } from 'react';
import {
  View,
  Text,
  TextInput,
  TouchableOpacity,
  FlatList,
  ActivityIndicator,
  StyleSheet,
  Alert,
} from 'react-native';
import { SafeAreaView } from 'react-native-safe-area-context';
import type { NativeStackScreenProps } from '@react-navigation/native-stack';
import type { ChatStackParamList } from '@/navigation/types';
import { colors, fontSize, fontFamily, spacing } from '@/theme';
import { chatsApi } from '@/services/api/chats';
import { topicsApi } from '@/services/api/topics';
import type { MemberInfo } from '@/types/chat';

type Props = NativeStackScreenProps<ChatStackParamList, 'CreateTopic'>;

const EMOJI_OPTIONS = [
  '📌', '💬', '📋', '🎯', '💡', '📊', '🔧', '📝',
  '🌾', '🏠', '💼', '📦', '🎨', '🔬', '⚡', '🌍',
];

export function CreateTopicScreen({ route, navigation }: Props) {
  const { chatId, chatType } = route.params;
  const isPersonal = chatType === 'personal';

  const [step, setStep] = useState(isPersonal ? 2 : 1);
  const [members, setMembers] = useState<MemberInfo[]>([]);
  const [selectedIds, setSelectedIds] = useState<string[]>([]);
  const [name, setName] = useState('');
  const [icon, setIcon] = useState('📌');
  const [description, setDescription] = useState('');
  const [loading, setLoading] = useState(false);
  const [loadingMembers, setLoadingMembers] = useState(false);

  useEffect(() => {
    if (!isPersonal) {
      loadGroupMembers();
    }
  }, []);

  const loadGroupMembers = async () => {
    setLoadingMembers(true);
    try {
      const res = await chatsApi.getGroupInfo(chatId);
      const groupMembers = res.data.data.members ?? [];
      setMembers(groupMembers);
    } catch {
      Alert.alert('Error', 'Gagal memuat anggota grup');
    } finally {
      setLoadingMembers(false);
    }
  };

  const toggleMember = useCallback((userId: string) => {
    setSelectedIds((prev) =>
      prev.includes(userId)
        ? prev.filter((id) => id !== userId)
        : [...prev, userId],
    );
  }, []);

  const selectAll = () => {
    setSelectedIds(members.map((m) => m.user.id));
  };

  const handleCreate = async () => {
    if (!name.trim()) {
      Alert.alert('Error', 'Nama topik wajib diisi');
      return;
    }

    setLoading(true);
    try {
      await topicsApi.create({
        name: name.trim(),
        icon,
        description: description.trim(),
        parentId: chatId,
        memberIds: isPersonal ? undefined : selectedIds,
      });
      navigation.goBack();
    } catch {
      Alert.alert('Error', 'Gagal membuat topik');
    } finally {
      setLoading(false);
    }
  };

  // Step 1: Select members (group only)
  if (step === 1) {
    return (
      <SafeAreaView style={styles.container}>
        <View style={styles.header}>
          <Text style={styles.title}>Pilih Anggota Topik</Text>
          <TouchableOpacity onPress={selectAll}>
            <Text style={styles.selectAll}>Pilih Semua</Text>
          </TouchableOpacity>
        </View>

        {loadingMembers ? (
          <View style={styles.center}>
            <ActivityIndicator color={colors.green} />
          </View>
        ) : (
          <FlatList
            data={members}
            keyExtractor={(item) => item.user.id}
            contentContainerStyle={styles.listContent}
            renderItem={({ item }) => {
              const selected = selectedIds.includes(item.user.id);
              return (
                <TouchableOpacity
                  style={styles.memberRow}
                  onPress={() => toggleMember(item.user.id)}
                >
                  <View style={styles.memberAvatar}>
                    <Text style={styles.avatarText}>{item.user.avatar || '👤'}</Text>
                  </View>
                  <View style={styles.memberInfo}>
                    <Text style={styles.memberName}>{item.user.name}</Text>
                    <Text style={styles.memberRole}>{item.role}</Text>
                  </View>
                  <View style={[styles.checkbox, selected && styles.checkboxSelected]}>
                    {selected && <Text style={styles.checkmark}>✓</Text>}
                  </View>
                </TouchableOpacity>
              );
            }}
          />
        )}

        <View style={styles.bottomBar}>
          <TouchableOpacity
            style={[styles.nextButton, selectedIds.length === 0 && styles.buttonDisabled]}
            onPress={() => setStep(2)}
            disabled={selectedIds.length === 0}
          >
            <Text style={styles.nextButtonText}>
              Lanjut ({selectedIds.length} dipilih)
            </Text>
          </TouchableOpacity>
        </View>
      </SafeAreaView>
    );
  }

  // Step 2: Topic details
  return (
    <SafeAreaView style={styles.container}>
      <View style={styles.detailsContainer}>
        {/* Icon picker */}
        <Text style={styles.label}>Ikon Topik</Text>
        <View style={styles.emojiGrid}>
          {EMOJI_OPTIONS.map((emoji) => (
            <TouchableOpacity
              key={emoji}
              style={[styles.emojiItem, icon === emoji && styles.emojiSelected]}
              onPress={() => setIcon(emoji)}
            >
              <Text style={styles.emojiText}>{emoji}</Text>
            </TouchableOpacity>
          ))}
        </View>

        {/* Name */}
        <Text style={styles.label}>Nama Topik</Text>
        <TextInput
          style={styles.input}
          value={name}
          onChangeText={setName}
          placeholder="Contoh: Pembagian Lahan"
          placeholderTextColor={colors.textMuted}
          maxLength={100}
          autoFocus
        />

        {/* Description */}
        <Text style={styles.label}>Deskripsi (opsional)</Text>
        <TextInput
          style={[styles.input, styles.textArea]}
          value={description}
          onChangeText={setDescription}
          placeholder="Deskripsi topik..."
          placeholderTextColor={colors.textMuted}
          multiline
          numberOfLines={3}
          textAlignVertical="top"
        />

        {isPersonal && (
          <Text style={styles.autoMemberNote}>
            Kedua peserta chat akan otomatis menjadi anggota topik ini.
          </Text>
        )}
      </View>

      <View style={styles.bottomBar}>
        {!isPersonal && (
          <TouchableOpacity
            style={styles.backButton}
            onPress={() => setStep(1)}
          >
            <Text style={styles.backButtonText}>Kembali</Text>
          </TouchableOpacity>
        )}
        <TouchableOpacity
          style={[
            styles.createButton,
            (!name.trim() || loading) && styles.buttonDisabled,
          ]}
          onPress={handleCreate}
          disabled={!name.trim() || loading}
        >
          {loading ? (
            <ActivityIndicator color={colors.white} size="small" />
          ) : (
            <Text style={styles.createButtonText}>Buat Topik</Text>
          )}
        </TouchableOpacity>
      </View>
    </SafeAreaView>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: colors.background,
  },
  header: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    paddingHorizontal: spacing.lg,
    paddingVertical: spacing.md,
    borderBottomWidth: 1,
    borderBottomColor: colors.border,
  },
  title: {
    fontFamily: fontFamily.ui,
    fontSize: fontSize.lg,
    fontWeight: '600',
    color: colors.textPrimary,
  },
  selectAll: {
    fontFamily: fontFamily.ui,
    fontSize: fontSize.sm,
    color: colors.green,
    fontWeight: '600',
  },
  center: {
    flex: 1,
    justifyContent: 'center',
    alignItems: 'center',
  },
  listContent: {
    paddingVertical: spacing.sm,
  },
  memberRow: {
    flexDirection: 'row',
    alignItems: 'center',
    paddingHorizontal: spacing.lg,
    paddingVertical: spacing.md,
  },
  memberAvatar: {
    width: 40,
    height: 40,
    borderRadius: 20,
    backgroundColor: colors.surface2,
    justifyContent: 'center',
    alignItems: 'center',
  },
  avatarText: {
    fontSize: fontSize.xl,
  },
  memberInfo: {
    flex: 1,
    marginLeft: spacing.md,
  },
  memberName: {
    fontFamily: fontFamily.ui,
    fontSize: fontSize.md,
    color: colors.textPrimary,
    fontWeight: '500',
  },
  memberRole: {
    fontFamily: fontFamily.ui,
    fontSize: fontSize.xs,
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
    color: colors.white,
    fontSize: fontSize.sm,
    fontWeight: '700',
  },
  bottomBar: {
    flexDirection: 'row',
    paddingHorizontal: spacing.lg,
    paddingVertical: spacing.md,
    borderTopWidth: 1,
    borderTopColor: colors.border,
    gap: spacing.md,
  },
  nextButton: {
    flex: 1,
    backgroundColor: colors.green,
    paddingVertical: spacing.md,
    borderRadius: spacing.sm,
    alignItems: 'center',
  },
  nextButtonText: {
    fontFamily: fontFamily.ui,
    fontSize: fontSize.md,
    color: colors.white,
    fontWeight: '600',
  },
  buttonDisabled: {
    opacity: 0.5,
  },
  detailsContainer: {
    flex: 1,
    paddingHorizontal: spacing.lg,
    paddingTop: spacing.lg,
  },
  label: {
    fontFamily: fontFamily.ui,
    fontSize: fontSize.sm,
    color: colors.textMuted,
    fontWeight: '600',
    marginBottom: spacing.sm,
    marginTop: spacing.lg,
  },
  emojiGrid: {
    flexDirection: 'row',
    flexWrap: 'wrap',
    gap: spacing.sm,
  },
  emojiItem: {
    width: 44,
    height: 44,
    borderRadius: spacing.sm,
    backgroundColor: colors.surface,
    justifyContent: 'center',
    alignItems: 'center',
  },
  emojiSelected: {
    backgroundColor: colors.surface2,
    borderWidth: 2,
    borderColor: colors.green,
  },
  emojiText: {
    fontSize: fontSize.xl,
  },
  input: {
    fontFamily: fontFamily.ui,
    fontSize: fontSize.md,
    color: colors.textPrimary,
    backgroundColor: colors.surface,
    borderRadius: spacing.sm,
    paddingHorizontal: spacing.md,
    paddingVertical: spacing.md,
    borderWidth: 1,
    borderColor: colors.border,
  },
  textArea: {
    height: 80,
    paddingTop: spacing.md,
  },
  autoMemberNote: {
    fontFamily: fontFamily.ui,
    fontSize: fontSize.sm,
    color: colors.textMuted,
    fontStyle: 'italic',
    marginTop: spacing.lg,
    textAlign: 'center',
  },
  backButton: {
    paddingVertical: spacing.md,
    paddingHorizontal: spacing.lg,
    borderRadius: spacing.sm,
    borderWidth: 1,
    borderColor: colors.border,
    alignItems: 'center',
  },
  backButtonText: {
    fontFamily: fontFamily.ui,
    fontSize: fontSize.md,
    color: colors.textPrimary,
    fontWeight: '500',
  },
  createButton: {
    flex: 1,
    backgroundColor: colors.green,
    paddingVertical: spacing.md,
    borderRadius: spacing.sm,
    alignItems: 'center',
  },
  createButtonText: {
    fontFamily: fontFamily.ui,
    fontSize: fontSize.md,
    color: colors.white,
    fontWeight: '600',
  },
});
