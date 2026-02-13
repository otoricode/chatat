// Chat info screen — group info with member management, or personal chat info
import React, { useEffect, useState, useCallback } from 'react';
import {
  View,
  Text,
  FlatList,
  Pressable,
  Alert,
  TextInput,
  ScrollView,
  ActivityIndicator,
  StyleSheet,
} from 'react-native';
import { SafeAreaView } from 'react-native-safe-area-context';
import type { NativeStackScreenProps } from '@react-navigation/native-stack';
import type { ChatStackParamList } from '@/navigation/types';
import { Avatar } from '@/components/ui/Avatar';
import { useAuthStore } from '@/stores/authStore';
import { useChatStore } from '@/stores/chatStore';
import { chatsApi } from '@/services/api/chats';
import { formatLastSeen } from '@/lib/timeFormat';
import { useTranslation } from 'react-i18next';
import { colors, fontSize, fontFamily, spacing } from '@/theme';
import type { MemberInfo, GroupInfo } from '@/types/chat';

type Props = NativeStackScreenProps<ChatStackParamList, 'ChatInfo'>;

export function ChatInfoScreen({ route, navigation }: Props) {
  const { chatId, chatType } = route.params;
  const { t } = useTranslation();
  const currentUserId = useAuthStore((s) => s.user?.id);
  const chatItem = useChatStore((s) => s.chats.find((c) => c.chat.id === chatId));
  const isGroup = chatType === 'group';

  const [groupInfo, setGroupInfo] = useState<GroupInfo | null>(null);
  const [loading, setLoading] = useState(true);
  const [isEditing, setIsEditing] = useState(false);
  const [editName, setEditName] = useState('');
  const [editDescription, setEditDescription] = useState('');

  const loadGroupInfo = useCallback(async () => {
    if (!isGroup) {
      setLoading(false);
      return;
    }
    try {
      const res = await chatsApi.getGroupInfo(chatId);
      setGroupInfo(res.data.data);
    } catch {
      Alert.alert(t('common.failed'), t('group.loadFailed'));
    } finally {
      setLoading(false);
    }
  }, [chatId, isGroup]);

  useEffect(() => {
    loadGroupInfo();
  }, [loadGroupInfo]);

  useEffect(() => {
    navigation.setOptions({
      title: isGroup ? t('group.groupInfo') : t('chat.contactInfo'),
      headerStyle: { backgroundColor: colors.headerBackground },
      headerTintColor: colors.textPrimary,
    });
  }, [navigation, isGroup]);

  const currentMember = groupInfo?.members.find((m) => m.user.id === currentUserId);
  const isAdmin = currentMember?.role === 'admin';
  const isCreator = groupInfo?.chat.createdBy === currentUserId;

  // --- Group Admin Actions ---

  const handleSaveEdit = async () => {
    if (!editName.trim()) {
      Alert.alert(t('group.nameRequired'));
      return;
    }
    try {
      await chatsApi.updateGroup(chatId, {
        name: editName.trim(),
        description: editDescription.trim(),
      });
      setIsEditing(false);
      loadGroupInfo();
    } catch {
      Alert.alert(t('common.failed'), t('group.updateFailed'));
    }
  };

  const handleAddMember = () => {
    // Navigate to contact list with selection mode
    // For now, prompt userId
    Alert.prompt?.(
      t('group.addMembers'),
      t('group.enterUserId'),
      async (userId: string) => {
        if (!userId.trim()) return;
        try {
          await chatsApi.addMember(chatId, userId.trim());
          loadGroupInfo();
        } catch {
          Alert.alert(t('common.failed'), t('group.addMemberFailed'));
        }
      },
    );
  };

  const handleRemoveMember = (member: MemberInfo) => {
    Alert.alert(
      t('group.removeMember'),
      t('group.removeMemberConfirm', { name: member.user.name }),
      [
        { text: t('common.cancel'), style: 'cancel' },
        {
          text: t('group.removeMember'),
          style: 'destructive',
          onPress: async () => {
            try {
              await chatsApi.removeMember(chatId, member.user.id);
              loadGroupInfo();
            } catch {
              Alert.alert(t('common.failed'), t('group.removeMemberFailed'));
            }
          },
        },
      ],
    );
  };

  const handlePromoteToAdmin = (member: MemberInfo) => {
    Alert.alert(
      t('group.promoteAdmin'),
      t('group.promoteAdminConfirm', { name: member.user.name }),
      [
        { text: t('common.cancel'), style: 'cancel' },
        {
          text: t('common.yes'),
          onPress: async () => {
            try {
              await chatsApi.promoteToAdmin(chatId, member.user.id);
              loadGroupInfo();
            } catch {
              Alert.alert(t('common.failed'), t('group.promoteAdminFailed'));
            }
          },
        },
      ],
    );
  };

  const handleLeaveGroup = () => {
    Alert.alert(t('group.leaveGroup'), t('group.areYouSure'), [
      { text: t('common.cancel'), style: 'cancel' },
      {
        text: t('group.leaveGroup'),
        style: 'destructive',
        onPress: async () => {
          try {
            await chatsApi.leaveGroup(chatId);
            navigation.popToTop();
          } catch {
            Alert.alert(t('common.failed'), t('group.leaveFailed'));
          }
        },
      },
    ]);
  };

  const handleDeleteGroup = () => {
    Alert.alert(
      t('group.deleteGroup'),
      t('group.allMessagesDeleted'),
      [
        { text: t('common.cancel'), style: 'cancel' },
        {
          text: t('common.delete'),
          style: 'destructive',
          onPress: async () => {
            try {
              await chatsApi.deleteGroup(chatId);
              navigation.popToTop();
            } catch {
              Alert.alert(t('common.failed'), t('group.deleteFailed'));
            }
          },
        },
      ],
    );
  };

  const handleMemberPress = (member: MemberInfo) => {
    if (!isAdmin || member.user.id === currentUserId) return;

    const options: Array<{
      text: string;
      onPress?: () => void;
      style?: 'cancel' | 'destructive';
    }> = [];

    if (member.role !== 'admin') {
      options.push({
        text: t('group.promoteAdmin'),
        onPress: () => handlePromoteToAdmin(member),
      });
    }

    options.push({
      text: t('group.removeMember'),
      style: 'destructive',
      onPress: () => handleRemoveMember(member),
    });

    options.push({ text: t('common.cancel'), style: 'cancel' });

    Alert.alert(member.user.name, undefined, options);
  };

  // --- RENDER ---

  if (loading) {
    return (
      <SafeAreaView style={styles.container}>
        <View style={styles.centered}>
          <ActivityIndicator color={colors.green} size="large" />
        </View>
      </SafeAreaView>
    );
  }

  // --- Personal Chat Info ---
  if (!isGroup) {
    const otherUser = chatItem?.otherUser;
    return (
      <SafeAreaView style={styles.container}>
        <ScrollView contentContainerStyle={styles.scrollContent}>
          <View style={styles.profileSection}>
            <Avatar
              emoji={otherUser?.avatar || '\u{1F464}'}
              size="lg"
              online={chatItem?.isOnline}
            />
            <Text style={styles.profileName}>{otherUser?.name || t('common.user')}</Text>
            <Text style={styles.profileStatus}>
              {otherUser
                ? formatLastSeen(otherUser.lastSeen, chatItem?.isOnline ?? false, t)
                : ''}
            </Text>
            <Text style={styles.profilePhone}>{otherUser?.phone || ''}</Text>
          </View>

          {otherUser?.status ? (
            <View style={styles.section}>
              <Text style={styles.sectionTitle}>{t('common.status')}</Text>
              <Text style={styles.sectionBody}>{otherUser.status}</Text>
            </View>
          ) : null}

          <View style={styles.section}>
            <Pressable
              style={styles.topicRow}
              onPress={() => navigation.navigate('TopicList', { chatId })}
            >
              <Text style={styles.topicIcon}>📌</Text>
              <Text style={styles.topicLabel}>{t('chat.discussionTopics')}</Text>
              <Text style={styles.topicArrow}>›</Text>
            </Pressable>
          </View>
        </ScrollView>
      </SafeAreaView>
    );
  }

  // --- Group Chat Info ---
  const chat = groupInfo?.chat;
  const members = groupInfo?.members ?? [];

  const renderMemberItem = ({ item }: { item: MemberInfo }) => {
    const isMe = item.user.id === currentUserId;
    return (
      <Pressable
        style={styles.memberRow}
        onPress={() => handleMemberPress(item)}
        disabled={!isAdmin || isMe}
      >
        <Avatar
          emoji={item.user.avatar || '\u{1F464}'}
          size="sm"
          online={item.isOnline}
        />
        <View style={styles.memberInfo}>
          <Text style={styles.memberName}>
            {item.user.name}
            {isMe ? ` ${t('common.you')}` : ''}
          </Text>
          <Text style={styles.memberRole}>
            {item.role === 'admin' ? t('group.admin') : t('group.member')}
          </Text>
        </View>
      </Pressable>
    );
  };

  return (
    <SafeAreaView style={styles.container}>
      <ScrollView contentContainerStyle={styles.scrollContent}>
        {/* Group profile */}
        <View style={styles.profileSection}>
          <Avatar emoji={chat?.icon || '\u{1F465}'} size="lg" />
          {isEditing ? (
            <View style={styles.editSection}>
              <TextInput
                style={styles.editInput}
                value={editName}
                onChangeText={setEditName}
                placeholder={t('group.groupNamePlaceholder')}
                placeholderTextColor={colors.textMuted}
                maxLength={100}
              />
              <TextInput
                style={[styles.editInput, styles.editMultiline]}
                value={editDescription}
                onChangeText={setEditDescription}
                placeholder={t('group.descriptionPlaceholder')}
                placeholderTextColor={colors.textMuted}
                multiline
                maxLength={500}
              />
              <View style={styles.editActions}>
                <Pressable
                  style={[styles.editBtn, styles.cancelBtn]}
                  onPress={() => setIsEditing(false)}
                >
                  <Text style={styles.cancelBtnText}>{t('common.cancel')}</Text>
                </Pressable>
                <Pressable
                  style={[styles.editBtn, styles.saveBtn]}
                  onPress={handleSaveEdit}
                >
                  <Text style={styles.saveBtnText}>{t('common.save')}</Text>
                </Pressable>
              </View>
            </View>
          ) : (
            <>
              <Text style={styles.profileName}>{chat?.name || t('group.groupInfo')}</Text>
              {chat?.description ? (
                <Text style={styles.profileStatus}>{chat.description}</Text>
              ) : null}
              <Text style={styles.memberCount}>{t('group.memberCount', { count: members.length })}</Text>

              {isAdmin && (
                <Pressable
                  style={styles.editProfileBtn}
                  onPress={() => {
                    setEditName(chat?.name || '');
                    setEditDescription(chat?.description || '');
                    setIsEditing(true);
                  }}
                >
                  <Text style={styles.editProfileBtnText}>{t('group.editGroupInfo')}</Text>
                </Pressable>
              )}
            </>
          )}
        </View>

        {/* Topics shortcut */}
        <View style={styles.section}>
          <Pressable
            style={styles.topicRow}
            onPress={() => navigation.navigate('TopicList', { chatId })}
          >
            <Text style={styles.topicIcon}>📌</Text>
              <Text style={styles.topicLabel}>{t('chat.discussionTopics')}</Text>
              <Text style={styles.topicArrow}>›</Text>
          </Pressable>
        </View>

        {/* Members section */}
        <View style={styles.section}>
          <View style={styles.sectionHeader}>
            <Text style={styles.sectionTitle}>
              {t('group.membersCount', { count: members.length })}
            </Text>
            {isAdmin && (
              <Pressable onPress={handleAddMember}>
                <Text style={styles.addMemberBtn}>+ {t('common.add')}</Text>
              </Pressable>
            )}
          </View>
          <FlatList
            data={members}
            renderItem={renderMemberItem}
            keyExtractor={(item) => item.user.id}
            scrollEnabled={false}
          />
        </View>

        {/* Actions */}
        <View style={styles.section}>
          {!isCreator && (
            <Pressable style={styles.dangerRow} onPress={handleLeaveGroup}>
              <Text style={styles.dangerText}>{t('group.leaveGroup')}</Text>
            </Pressable>
          )}
          {isCreator && (
            <Pressable style={styles.dangerRow} onPress={handleDeleteGroup}>
              <Text style={styles.dangerText}>{t('group.deleteGroup')}</Text>
            </Pressable>
          )}
        </View>
      </ScrollView>
    </SafeAreaView>
  );
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
  scrollContent: {
    paddingBottom: spacing.xl,
  },
  profileSection: {
    alignItems: 'center',
    paddingVertical: spacing.xl,
    paddingHorizontal: spacing.lg,
    borderBottomWidth: 1,
    borderBottomColor: colors.surface2,
  },
  profileName: {
    fontFamily: fontFamily.uiSemiBold,
    fontSize: fontSize.xl,
    color: colors.textPrimary,
    marginTop: spacing.md,
    textAlign: 'center',
  },
  profileStatus: {
    fontFamily: fontFamily.ui,
    fontSize: fontSize.sm,
    color: colors.textMuted,
    marginTop: spacing.xs,
    textAlign: 'center',
  },
  profilePhone: {
    fontFamily: fontFamily.ui,
    fontSize: fontSize.sm,
    color: colors.textMuted,
    marginTop: spacing.xs,
  },
  memberCount: {
    fontFamily: fontFamily.ui,
    fontSize: fontSize.sm,
    color: colors.textMuted,
    marginTop: spacing.xs,
  },
  section: {
    paddingHorizontal: spacing.lg,
    paddingVertical: spacing.md,
    borderBottomWidth: 1,
    borderBottomColor: colors.surface2,
  },
  sectionHeader: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    marginBottom: spacing.sm,
  },
  sectionTitle: {
    fontFamily: fontFamily.uiSemiBold,
    fontSize: fontSize.sm,
    color: colors.textMuted,
    textTransform: 'uppercase',
    letterSpacing: 0.5,
  },
  sectionBody: {
    fontFamily: fontFamily.ui,
    fontSize: fontSize.md,
    color: colors.textPrimary,
    marginTop: spacing.xs,
  },
  addMemberBtn: {
    fontFamily: fontFamily.uiSemiBold,
    fontSize: fontSize.sm,
    color: colors.green,
  },
  memberRow: {
    flexDirection: 'row',
    alignItems: 'center',
    paddingVertical: spacing.sm,
    gap: spacing.md,
  },
  memberInfo: {
    flex: 1,
  },
  memberName: {
    fontFamily: fontFamily.ui,
    fontSize: fontSize.md,
    color: colors.textPrimary,
  },
  memberRole: {
    fontFamily: fontFamily.ui,
    fontSize: fontSize.xs,
    color: colors.textMuted,
  },
  editSection: {
    width: '100%',
    marginTop: spacing.md,
    gap: spacing.sm,
  },
  editInput: {
    fontFamily: fontFamily.ui,
    fontSize: fontSize.md,
    color: colors.textPrimary,
    backgroundColor: colors.surface2,
    borderRadius: 10,
    paddingHorizontal: spacing.md,
    paddingVertical: spacing.sm,
  },
  editMultiline: {
    minHeight: 60,
    textAlignVertical: 'top',
  },
  editActions: {
    flexDirection: 'row',
    gap: spacing.sm,
    justifyContent: 'flex-end',
  },
  editBtn: {
    paddingHorizontal: spacing.lg,
    paddingVertical: spacing.sm,
    borderRadius: 8,
  },
  cancelBtn: {
    backgroundColor: colors.surface2,
  },
  cancelBtnText: {
    fontFamily: fontFamily.uiSemiBold,
    fontSize: fontSize.sm,
    color: colors.textMuted,
  },
  saveBtn: {
    backgroundColor: colors.green,
  },
  saveBtnText: {
    fontFamily: fontFamily.uiSemiBold,
    fontSize: fontSize.sm,
    color: colors.background,
  },
  editProfileBtn: {
    marginTop: spacing.md,
    paddingHorizontal: spacing.lg,
    paddingVertical: spacing.sm,
    backgroundColor: colors.surface2,
    borderRadius: 8,
  },
  editProfileBtnText: {
    fontFamily: fontFamily.uiSemiBold,
    fontSize: fontSize.sm,
    color: colors.green,
  },
  dangerRow: {
    paddingVertical: spacing.md,
  },
  dangerText: {
    fontFamily: fontFamily.uiSemiBold,
    fontSize: fontSize.md,
    color: '#EF4444',
  },
  topicRow: {
    flexDirection: 'row',
    alignItems: 'center',
    paddingVertical: spacing.sm,
  },
  topicIcon: {
    fontSize: fontSize.xl,
    marginRight: spacing.md,
  },
  topicLabel: {
    fontFamily: fontFamily.ui,
    fontSize: fontSize.md,
    color: colors.textPrimary,
    flex: 1,
  },
  topicArrow: {
    fontFamily: fontFamily.ui,
    fontSize: fontSize.xl,
    color: colors.textMuted,
  },
});
