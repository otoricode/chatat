// Document editor screen — Notion-style block editor
import React, { useCallback, useEffect, useState } from 'react';
import {
  View,
  Text,
  TextInput,
  Pressable,
  StyleSheet,
  ActivityIndicator,
  Alert,
} from 'react-native';
import { SafeAreaView } from 'react-native-safe-area-context';
import type { NativeStackScreenProps } from '@react-navigation/native-stack';
import type { DocumentStackParamList } from '@/navigation/types';
import { colors, fontSize, fontFamily, spacing } from '@/theme';
import { useEditorStore } from '@/stores/editorStore';
import { BlockEditor } from '@/components/editor';
import { SaveIndicator } from '@/components/editor/SaveIndicator';
import { LockStatusBadge } from '@/components/document/LockStatusBadge';
import { LockActionSheet } from '@/components/document/LockActionSheet';
import { SignConfirmModal } from '@/components/document/SignConfirmModal';
import { documentsApi } from '@/services/api/documents';

type Props = NativeStackScreenProps<DocumentStackParamList, 'DocumentEditor'>;

export function DocumentEditorScreen({ route, navigation }: Props) {
  const documentId = route.params?.documentId;
  const contextType = route.params?.contextType;
  const contextId = route.params?.contextId;

  const title = useEditorStore((s) => s.title);
  const icon = useEditorStore((s) => s.icon);
  const isLoading = useEditorStore((s) => s.isLoading);
  const isLocked = useEditorStore((s) => s.isLocked);
  const error = useEditorStore((s) => s.error);

  const loadDocument = useEditorStore((s) => s.loadDocument);
  const createDocument = useEditorStore((s) => s.createDocument);
  const updateTitle = useEditorStore((s) => s.updateTitle);
  const updateIcon = useEditorStore((s) => s.updateIcon);
  const reset = useEditorStore((s) => s.reset);

  const [initialized, setInitialized] = useState(false);
  const [showLockSheet, setShowLockSheet] = useState(false);
  const [showSignModal, setShowSignModal] = useState(false);
  const [lockLoading, setLockLoading] = useState(false);
  const [lockedBy, setLockedBy] = useState<string | null>(null);

  useEffect(() => {
    const init = async () => {
      if (documentId) {
        await loadDocument(documentId);
      } else {
        const chatId = contextType === 'chat' ? contextId : undefined;
        const topicId = contextType === 'topic' ? contextId : undefined;
        await createDocument('Dokumen Baru', chatId, topicId);
      }
      setInitialized(true);
    };
    init();

    return () => {
      reset();
    };
  }, [documentId, contextType, contextId, loadDocument, createDocument, reset]);

  // Fetch lock info when document loads
  useEffect(() => {
    if (documentId && initialized) {
      documentsApi.getById(documentId).then((res) => {
        const doc = res.data.data;
        setLockedBy(doc.document.lockedBy ?? null);
      }).catch(() => {});
    }
  }, [documentId, initialized, isLocked]);

  const handleBack = () => {
    navigation.goBack();
  };

  const handleLockManual = useCallback(async () => {
    const docId = useEditorStore.getState().documentId;
    if (!docId) return;
    setLockLoading(true);
    try {
      await documentsApi.lock(docId, 'manual');
      await loadDocument(docId);
      setLockedBy('manual');
      setShowLockSheet(false);
    } catch {
      Alert.alert('Gagal', 'Tidak dapat mengunci dokumen');
    }
    setLockLoading(false);
  }, [loadDocument]);

  const handleLockSignatures = useCallback(async () => {
    const docId = useEditorStore.getState().documentId;
    if (!docId) return;
    setLockLoading(true);
    try {
      await documentsApi.lock(docId, 'signatures');
      await loadDocument(docId);
      setLockedBy('signatures');
      setShowLockSheet(false);
    } catch {
      Alert.alert('Gagal', 'Tambahkan penandatangan terlebih dahulu');
    }
    setLockLoading(false);
  }, [loadDocument]);

  const handleUnlock = useCallback(async () => {
    const docId = useEditorStore.getState().documentId;
    if (!docId) return;
    setLockLoading(true);
    try {
      await documentsApi.unlock(docId);
      await loadDocument(docId);
      setLockedBy(null);
      setShowLockSheet(false);
    } catch {
      Alert.alert('Gagal', 'Tidak dapat membuka kunci dokumen');
    }
    setLockLoading(false);
  }, [loadDocument]);

  const handleSign = useCallback(async (name: string) => {
    const docId = useEditorStore.getState().documentId;
    if (!docId) return;
    setLockLoading(true);
    try {
      await documentsApi.sign(docId, name);
      await loadDocument(docId);
      setShowSignModal(false);
    } catch {
      Alert.alert('Gagal', 'Tidak dapat menandatangani dokumen');
    }
    setLockLoading(false);
  }, [loadDocument]);

  if (isLoading && !initialized) {
    return (
      <SafeAreaView style={styles.container}>
        <View style={styles.loading}>
          <ActivityIndicator size="large" color={colors.green} />
        </View>
      </SafeAreaView>
    );
  }

  if (error) {
    return (
      <SafeAreaView style={styles.container}>
        <View style={styles.header}>
          <Pressable onPress={handleBack} style={styles.backBtn}>
            <Text style={styles.backText}>←</Text>
          </Pressable>
        </View>
        <View style={styles.loading}>
          <Text style={styles.errorText}>{error}</Text>
        </View>
      </SafeAreaView>
    );
  }

  return (
    <SafeAreaView style={styles.container} edges={['top']}>
      {/* Header */}
      <View style={styles.header}>
        <Pressable onPress={handleBack} style={styles.backBtn}>
          <Text style={styles.backText}>←</Text>
        </Pressable>
        <View style={styles.headerCenter}>
          <SaveIndicator />
        </View>
        <Pressable onPress={() => setShowLockSheet(true)}>
          <LockStatusBadge locked={isLocked} lockedBy={lockedBy} compact />
        </Pressable>
        {isLocked && lockedBy === 'signatures' && (
          <Pressable onPress={() => setShowSignModal(true)} style={styles.signBtn}>
            <Text style={styles.signBtnText}>Tandatangani</Text>
          </Pressable>
        )}
      </View>

      {/* Title area */}
      <View style={styles.titleArea}>
        <Pressable onPress={() => updateIcon(icon === '📄' ? '📝' : '📄')}>
          <Text style={styles.icon}>{icon}</Text>
        </Pressable>
        <TextInput
          style={styles.titleInput}
          value={title}
          onChangeText={updateTitle}
          placeholder="Judul dokumen"
          placeholderTextColor={colors.textMuted}
          editable={!isLocked}
          multiline
          blurOnSubmit
        />
      </View>

      {/* Editor */}
      <BlockEditor readOnly={isLocked} />

      {/* Lock Action Sheet */}
      <LockActionSheet
        visible={showLockSheet}
        locked={isLocked}
        lockedBy={lockedBy}
        loading={lockLoading}
        onLockManual={handleLockManual}
        onLockSignatures={handleLockSignatures}
        onUnlock={handleUnlock}
        onClose={() => setShowLockSheet(false)}
      />

      {/* Sign Confirmation */}
      <SignConfirmModal
        visible={showSignModal}
        loading={lockLoading}
        onSign={handleSign}
        onClose={() => setShowSignModal(false)}
      />
    </SafeAreaView>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: colors.background,
  },
  loading: {
    flex: 1,
    justifyContent: 'center',
    alignItems: 'center',
  },
  header: {
    flexDirection: 'row',
    alignItems: 'center',
    paddingHorizontal: spacing.md,
    paddingVertical: spacing.sm,
    borderBottomWidth: 1,
    borderBottomColor: colors.border,
  },
  backBtn: {
    padding: spacing.sm,
  },
  backText: {
    fontSize: fontSize.xl,
    color: colors.textPrimary,
  },
  headerCenter: {
    flex: 1,
    alignItems: 'center',
  },
  signBtn: {
    backgroundColor: colors.green,
    paddingHorizontal: spacing.md,
    paddingVertical: spacing.xs,
    borderRadius: 6,
    marginLeft: spacing.sm,
  },
  signBtnText: {
    fontFamily: fontFamily.uiSemiBold,
    fontSize: fontSize.xs,
    color: colors.background,
  },
  titleArea: {
    flexDirection: 'row',
    alignItems: 'flex-start',
    paddingHorizontal: spacing.lg,
    paddingVertical: spacing.md,
  },
  icon: {
    fontSize: 28,
    marginRight: spacing.sm,
    marginTop: 2,
  },
  titleInput: {
    flex: 1,
    fontFamily: fontFamily.documentBold,
    fontSize: fontSize.h2,
    color: colors.textPrimary,
    lineHeight: 32,
    padding: 0,
  },
  errorText: {
    fontFamily: fontFamily.ui,
    fontSize: fontSize.md,
    color: colors.red,
  },
});
