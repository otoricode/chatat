// Topic info screen — details, members, admin actions
import React, { useEffect, useState, useCallback } from 'react';
import {
  View,
  Text,
  FlatList,
  TouchableOpacity,
  TextInput,
  ActivityIndicator,
  Alert,
  StyleSheet,
  ScrollView,
} from 'react-native';
import { SafeAreaView } from 'react-native-safe-area-context';
import type { NativeStackScreenProps } from '@react-navigation/native-stack';
import type { ChatStackParamList } from '@/navigation/types';
import { Avatar } from '@/components/ui/Avatar';
import { topicsApi } from '@/services/api/topics';
import { useAuthStore } from '@/stores/authStore';
import { useTopicStore } from '@/stores/topicStore';
import { useTranslation } from 'react-i18next';
import { colors, fontSize, fontFamily, spacing } from '@/theme';
import type { TopicDetail, MemberInfo } from '@/types/chat';

type Props = NativeStackScreenProps<ChatStackParamList, 'TopicInfo'>;

export function TopicInfoScreen({ route, navigation }: Props) {
  const { topicId } = route.params;
  const { t } = useTranslation();
  const currentUserId = useAuthStore((s) => s.user?.id);

  const [detail, setDetail] = useState<TopicDetail | null>(null);
  const [loading, setLoading] = useState(true);
  const [editing, setEditing] = useState(false);
  const [editName, setEditName] = useState('');
  const [editDesc, setEditDesc] = useState('');
  const [saving, setSaving] = useState(false);

  const isAdmin = detail?.members.some(
    (m) => m.user.id === currentUserId && m.role === 'admin',
  );

  useEffect(() => {
    loadDetail();
  }, [topicId]);

  const loadDetail = async () => {
    setLoading(true);
    try {
      const res = await topicsApi.getById(topicId);
      setDetail(res.data.data);
      setEditName(res.data.data.topic.name);
      setEditDesc(res.data.data.topic.description);
    } catch {
      Alert.alert(t('common.error'), t('topic.loadFailed'));
    } finally {
      setLoading(false);
    }
  };

  const handleSave = async () => {
    setSaving(true);
    try {
      await topicsApi.update(topicId, {
        name: editName.trim(),
        description: editDesc.trim(),
      });
      setEditing(false);
      loadDetail();
    } catch {
      Alert.alert(t('common.error'), t('topic.saveFailed'));
    } finally {
      setSaving(false);
    }
  };

  const handleRemoveMember = useCallback(
    (userId: string, userName: string) => {
      Alert.alert(
        t('topic.removeMember'),
        t('topic.removeMemberConfirm', { name: userName }),
        [
          { text: t('common.cancel'), style: 'cancel' },
          {
            text: t('common.delete'),
            style: 'destructive',
            onPress: async () => {
              try {
                await topicsApi.removeMember(topicId, userId);
                loadDetail();
              } catch {
                Alert.alert(t('common.error'), t('topic.removeMemberFailed'));
              }
            },
          },
        ],
      );
    },
    [topicId],
  );

  const handleDelete = () => {
    Alert.alert(
      t('topic.deleteTopic'),
      t('topic.deleteTopicConfirm'),
      [
        { text: t('common.cancel'), style: 'cancel' },
        {
          text: t('common.delete'),
          style: 'destructive',
          onPress: async () => {
            try {
              await topicsApi.delete(topicId);
              // Also remove from store
              if (detail?.topic.parentId) {
                useTopicStore.getState().removeTopic(detail.topic.parentId, topicId);
              }
              // Go back 2 screens (TopicScreen -> TopicList)
              navigation.pop(2);
            } catch {
              Alert.alert(t('common.error'), t('topic.deleteFailed'));
            }
          },
        },
      ],
    );
  };

  if (loading) {
    return (
      <SafeAreaView style={styles.container}>
        <View style={styles.center}>
          <ActivityIndicator color={colors.green} />
        </View>
      </SafeAreaView>
    );
  }

  if (!detail) return null;

  const { topic, members, parent } = detail;

  const renderMember = ({ item }: { item: MemberInfo }) => (
    <View style={styles.memberRow}>
      <Avatar emoji={item.user.avatar || '\u{1F464}'} size="sm" online={item.isOnline} />
      <View style={styles.memberInfo}>
        <Text style={styles.memberName}>
          {item.user.name}
          {item.user.id === currentUserId ? ` ${t('common.you')}` : ''}
        </Text>
        {item.role === 'admin' && (
          <Text style={styles.adminBadge}>Admin</Text>
        )}
      </View>
      {isAdmin && item.user.id !== currentUserId && (
        <TouchableOpacity
          onPress={() => handleRemoveMember(item.user.id, item.user.name)}
          style={styles.removeButton}
        >
          <Text style={styles.removeText}>{t('common.remove')}</Text>
        </TouchableOpacity>
      )}
    </View>
  );

  return (
    <SafeAreaView style={styles.container} edges={['bottom']}>
      <ScrollView contentContainerStyle={styles.scrollContent}>
        {/* Topic icon & name */}
        <View style={styles.topSection}>
          <View style={styles.bigIcon}>
            <Text style={styles.bigIconText}>{topic.icon || '📌'}</Text>
          </View>

          {editing ? (
            <View style={styles.editSection}>
              <TextInput
                style={styles.editInput}
                value={editName}
                onChangeText={setEditName}
                placeholder={t('topic.topicName')}
                placeholderTextColor={colors.textMuted}
              />
              <TextInput
                style={[styles.editInput, styles.editTextArea]}
                value={editDesc}
                onChangeText={setEditDesc}
                placeholder={t('topic.descriptionOptional')}
                placeholderTextColor={colors.textMuted}
                multiline
                textAlignVertical="top"
              />
              <View style={styles.editActions}>
                <TouchableOpacity
                  onPress={() => setEditing(false)}
                  style={styles.cancelButton}
                >
                  <Text style={styles.cancelText}>{t('common.cancel')}</Text>
                </TouchableOpacity>
                <TouchableOpacity
                  onPress={handleSave}
                  style={styles.saveButton}
                  disabled={saving}
                >
                  {saving ? (
                    <ActivityIndicator color={colors.white} size="small" />
                  ) : (
                    <Text style={styles.saveText}>{t('common.save')}</Text>
                  )}
                </TouchableOpacity>
              </View>
            </View>
          ) : (
            <>
              <Text style={styles.topicName}>{topic.name}</Text>
              {topic.description ? (
                <Text style={styles.topicDesc}>{topic.description}</Text>
              ) : null}
              {isAdmin && (
                <TouchableOpacity onPress={() => setEditing(true)}>
                  <Text style={styles.editLink}>{t('common.edit')}</Text>
                </TouchableOpacity>
              )}
            </>
          )}
        </View>

        {/* Parent chat info */}
        {parent && (
          <View style={styles.section}>
            <Text style={styles.sectionTitle}>{t('topic.parentChat')}</Text>
            <View style={styles.parentRow}>
              <Avatar emoji={parent.icon || '\u{1F4AC}'} size="sm" />
              <Text style={styles.parentName}>{parent.name || 'Chat'}</Text>
            </View>
          </View>
        )}

        {/* Members */}
        <View style={styles.section}>
          <Text style={styles.sectionTitle}>
            {t('group.membersCount', { count: members.length })}
          </Text>
          {members.map((m) => (
            <React.Fragment key={m.user.id}>
              {renderMember({ item: m })}
            </React.Fragment>
          ))}
        </View>

        {/* Delete */}
        {isAdmin && (
          <View style={styles.dangerSection}>
            <TouchableOpacity style={styles.deleteButton} onPress={handleDelete}>
              <Text style={styles.deleteText}>{t('topic.deleteTopic')}</Text>
            </TouchableOpacity>
          </View>
        )}
      </ScrollView>
    </SafeAreaView>
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
  scrollContent: {
    paddingBottom: spacing.xxxl,
  },
  topSection: {
    alignItems: 'center',
    paddingVertical: spacing.xl,
    borderBottomWidth: 1,
    borderBottomColor: colors.border,
  },
  bigIcon: {
    width: 72,
    height: 72,
    borderRadius: spacing.md,
    backgroundColor: colors.surface2,
    justifyContent: 'center',
    alignItems: 'center',
    marginBottom: spacing.md,
  },
  bigIconText: {
    fontSize: 40,
  },
  topicName: {
    fontFamily: fontFamily.uiSemiBold,
    fontSize: fontSize.xl,
    color: colors.textPrimary,
    textAlign: 'center',
  },
  topicDesc: {
    fontFamily: fontFamily.ui,
    fontSize: fontSize.sm,
    color: colors.textMuted,
    textAlign: 'center',
    marginTop: spacing.xs,
    paddingHorizontal: spacing.xl,
  },
  editLink: {
    fontFamily: fontFamily.ui,
    fontSize: fontSize.sm,
    color: colors.green,
    fontWeight: '600',
    marginTop: spacing.sm,
  },
  editSection: {
    width: '100%',
    paddingHorizontal: spacing.lg,
    gap: spacing.sm,
  },
  editInput: {
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
  editTextArea: {
    height: 64,
    paddingTop: spacing.md,
  },
  editActions: {
    flexDirection: 'row',
    gap: spacing.sm,
    justifyContent: 'flex-end',
    marginTop: spacing.xs,
  },
  cancelButton: {
    paddingVertical: spacing.sm,
    paddingHorizontal: spacing.lg,
    borderRadius: spacing.sm,
    borderWidth: 1,
    borderColor: colors.border,
  },
  cancelText: {
    fontFamily: fontFamily.ui,
    fontSize: fontSize.sm,
    color: colors.textPrimary,
  },
  saveButton: {
    paddingVertical: spacing.sm,
    paddingHorizontal: spacing.lg,
    borderRadius: spacing.sm,
    backgroundColor: colors.green,
  },
  saveText: {
    fontFamily: fontFamily.ui,
    fontSize: fontSize.sm,
    color: colors.white,
    fontWeight: '600',
  },
  section: {
    paddingTop: spacing.lg,
    paddingHorizontal: spacing.lg,
    borderBottomWidth: 1,
    borderBottomColor: colors.border,
    paddingBottom: spacing.lg,
  },
  sectionTitle: {
    fontFamily: fontFamily.uiSemiBold,
    fontSize: fontSize.sm,
    color: colors.textMuted,
    textTransform: 'uppercase',
    letterSpacing: 0.5,
    marginBottom: spacing.md,
  },
  parentRow: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: spacing.md,
  },
  parentName: {
    fontFamily: fontFamily.ui,
    fontSize: fontSize.md,
    color: colors.textPrimary,
  },
  memberRow: {
    flexDirection: 'row',
    alignItems: 'center',
    paddingVertical: spacing.sm,
  },
  memberInfo: {
    flex: 1,
    marginLeft: spacing.md,
    flexDirection: 'row',
    alignItems: 'center',
    gap: spacing.sm,
  },
  memberName: {
    fontFamily: fontFamily.ui,
    fontSize: fontSize.md,
    color: colors.textPrimary,
  },
  adminBadge: {
    fontFamily: fontFamily.ui,
    fontSize: fontSize.xs,
    color: colors.green,
    backgroundColor: colors.surface2,
    paddingHorizontal: spacing.sm,
    paddingVertical: 2,
    borderRadius: 4,
    overflow: 'hidden',
  },
  removeButton: {
    paddingVertical: spacing.xs,
    paddingHorizontal: spacing.md,
  },
  removeText: {
    fontFamily: fontFamily.ui,
    fontSize: fontSize.sm,
    color: colors.red,
  },
  dangerSection: {
    paddingHorizontal: spacing.lg,
    paddingVertical: spacing.xl,
  },
  deleteButton: {
    backgroundColor: colors.surface,
    borderWidth: 1,
    borderColor: colors.red,
    paddingVertical: spacing.md,
    borderRadius: spacing.sm,
    alignItems: 'center',
  },
  deleteText: {
    fontFamily: fontFamily.uiSemiBold,
    fontSize: fontSize.md,
    color: colors.red,
  },
});
