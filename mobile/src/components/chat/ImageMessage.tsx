// ImageMessage — image bubble in chat with thumbnail and progress
import React, { useState } from 'react';
import {
  View,
  Image,
  Text,
  Pressable,
  StyleSheet,
  ActivityIndicator,
} from 'react-native';
import { colors, fontSize, fontFamily, spacing } from '@/theme';
import type { MediaResponse } from '@/types/chat';

type Props = {
  media: MediaResponse;
  isSelf: boolean;
  onPress?: () => void;
  uploadProgress?: number; // 0-100, undefined = not uploading
};

export function ImageMessage({ media, isSelf, onPress, uploadProgress }: Props) {
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(false);

  const imageUrl = media.thumbnailURL || media.url;
  const aspectRatio =
    media.width && media.height ? media.width / media.height : 4 / 3;
  const maxWidth = 250;
  const displayWidth = Math.min(maxWidth, media.width || maxWidth);
  const displayHeight = displayWidth / aspectRatio;

  const isUploading = typeof uploadProgress === 'number' && uploadProgress < 100;

  return (
    <Pressable
      style={[
        styles.container,
        isSelf ? styles.selfContainer : styles.otherContainer,
      ]}
      onPress={onPress}
    >
      <View
        style={[
          styles.imageWrapper,
          isSelf ? styles.selfBubble : styles.otherBubble,
          { width: displayWidth, height: displayHeight },
        ]}
      >
        {!error ? (
          <Image
            source={{ uri: imageUrl }}
            style={styles.image}
            resizeMode="cover"
            onLoadStart={() => setLoading(true)}
            onLoadEnd={() => setLoading(false)}
            onError={() => {
              setError(true);
              setLoading(false);
            }}
          />
        ) : (
          <View style={styles.errorContainer}>
            <Text style={styles.errorIcon}>{'\u26A0'}</Text>
            <Text style={styles.errorText}>Gagal memuat</Text>
          </View>
        )}

        {(loading || isUploading) && (
          <View style={styles.loadingOverlay}>
            <ActivityIndicator color={colors.white} />
            {isUploading && (
              <Text style={styles.progressText}>{uploadProgress}%</Text>
            )}
          </View>
        )}
      </View>
    </Pressable>
  );
}

const styles = StyleSheet.create({
  container: {
    paddingHorizontal: spacing.sm,
    paddingVertical: 2,
  },
  selfContainer: {
    alignItems: 'flex-end',
  },
  otherContainer: {
    alignItems: 'flex-start',
  },
  imageWrapper: {
    borderRadius: 12,
    overflow: 'hidden',
    maxWidth: 250,
  },
  selfBubble: {
    borderBottomRightRadius: 4,
  },
  otherBubble: {
    borderBottomLeftRadius: 4,
  },
  image: {
    width: '100%',
    height: '100%',
  },
  loadingOverlay: {
    ...StyleSheet.absoluteFillObject,
    backgroundColor: 'rgba(0,0,0,0.4)',
    justifyContent: 'center',
    alignItems: 'center',
    gap: spacing.xs,
  },
  progressText: {
    fontFamily: fontFamily.uiSemiBold,
    fontSize: fontSize.sm,
    color: colors.white,
  },
  errorContainer: {
    flex: 1,
    justifyContent: 'center',
    alignItems: 'center',
    backgroundColor: colors.surface2,
  },
  errorIcon: {
    fontSize: 24,
    marginBottom: spacing.xs,
  },
  errorText: {
    fontFamily: fontFamily.ui,
    fontSize: fontSize.xs,
    color: colors.textMuted,
  },
});
