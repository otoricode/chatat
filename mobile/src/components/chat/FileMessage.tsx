// FileMessage — file card in chat bubble
import React from 'react';
import { View, Text, Pressable, StyleSheet, ActivityIndicator } from 'react-native';
import { colors, fontSize, fontFamily, spacing } from '@/theme';
import type { MediaResponse } from '@/types/chat';

type Props = {
  media: MediaResponse;
  isSelf: boolean;
  onPress?: () => void;
  uploadProgress?: number;
};

function getFileIcon(contentType: string): string {
  if (contentType.includes('pdf')) return '\u{1F4C4}';
  if (contentType.includes('word') || contentType.includes('document'))
    return '\u{1F4DD}';
  if (contentType.includes('excel') || contentType.includes('spreadsheet'))
    return '\u{1F4CA}';
  if (contentType.includes('powerpoint') || contentType.includes('presentation'))
    return '\u{1F4CA}';
  if (contentType.includes('zip')) return '\u{1F4E6}';
  if (contentType.includes('text')) return '\u{1F4C3}';
  return '\u{1F4CE}';
}

function formatFileSize(bytes: number): string {
  if (bytes < 1024) return bytes + ' B';
  if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + ' KB';
  return (bytes / (1024 * 1024)).toFixed(1) + ' MB';
}

export function FileMessage({ media, isSelf, onPress, uploadProgress }: Props) {
  const isUploading = typeof uploadProgress === 'number' && uploadProgress < 100;

  return (
    <View style={[styles.wrapper, isSelf ? styles.selfWrapper : styles.otherWrapper]}>
      <Pressable
        style={({ pressed }) => [
          styles.container,
          isSelf ? styles.selfBubble : styles.otherBubble,
          pressed && styles.pressed,
        ]}
        onPress={onPress}
      >
        <View style={styles.iconContainer}>
          <Text style={styles.icon}>{getFileIcon(media.contentType)}</Text>
        </View>
        <View style={styles.info}>
          <Text style={styles.filename} numberOfLines={1}>
            {media.filename}
          </Text>
          <Text style={styles.meta}>
            {formatFileSize(media.size)}
            {media.contentType.split('/').pop()
              ? ` \u2022 ${media.contentType.split('/').pop()?.toUpperCase()}`
              : ''}
          </Text>
        </View>
        {isUploading && (
          <View style={styles.progress}>
            <ActivityIndicator size="small" color={colors.green} />
            <Text style={styles.progressText}>{uploadProgress}%</Text>
          </View>
        )}
      </Pressable>
    </View>
  );
}

const styles = StyleSheet.create({
  wrapper: {
    paddingHorizontal: spacing.sm,
    paddingVertical: 2,
  },
  selfWrapper: {
    alignItems: 'flex-end',
  },
  otherWrapper: {
    alignItems: 'flex-start',
  },
  container: {
    flexDirection: 'row',
    alignItems: 'center',
    maxWidth: '80%',
    paddingHorizontal: spacing.md,
    paddingVertical: spacing.sm,
    borderRadius: 12,
    gap: spacing.sm,
  },
  selfBubble: {
    backgroundColor: colors.bubbleSelf,
    borderBottomRightRadius: 4,
  },
  otherBubble: {
    backgroundColor: colors.bubbleOther,
    borderBottomLeftRadius: 4,
  },
  pressed: {
    opacity: 0.8,
  },
  iconContainer: {
    width: 40,
    height: 40,
    borderRadius: 8,
    backgroundColor: 'rgba(255,255,255,0.1)',
    justifyContent: 'center',
    alignItems: 'center',
  },
  icon: {
    fontSize: 20,
  },
  info: {
    flex: 1,
    gap: 2,
  },
  filename: {
    fontFamily: fontFamily.uiSemiBold,
    fontSize: fontSize.sm,
    color: colors.textPrimary,
  },
  meta: {
    fontFamily: fontFamily.ui,
    fontSize: fontSize.xs,
    color: colors.textMuted,
  },
  progress: {
    alignItems: 'center',
    gap: 2,
  },
  progressText: {
    fontFamily: fontFamily.ui,
    fontSize: 10,
    color: colors.textMuted,
  },
});
