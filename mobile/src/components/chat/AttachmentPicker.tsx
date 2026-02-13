// AttachmentPicker — bottom sheet for selecting media from camera, gallery, or files
import React from 'react';
import { View, Text, Pressable, StyleSheet, Modal } from 'react-native';
import * as ImagePicker from 'expo-image-picker';
import * as DocumentPicker from 'expo-document-picker';
import { colors, fontSize, fontFamily, spacing } from '@/theme';

type PickedMedia = {
  uri: string;
  filename: string;
  mimeType: string;
  size?: number;
  width?: number;
  height?: number;
  type: 'image' | 'file';
};

type Props = {
  visible: boolean;
  onClose: () => void;
  onPick: (media: PickedMedia) => void;
};

const OPTIONS = [
  { key: 'camera', icon: '\u{1F4F7}', label: 'Kamera' },
  { key: 'gallery', icon: '\u{1F5BC}', label: 'Galeri' },
  { key: 'file', icon: '\u{1F4CE}', label: 'File' },
];

export function AttachmentPicker({ visible, onClose, onPick }: Props) {
  const handleCamera = async () => {
    const permission = await ImagePicker.requestCameraPermissionsAsync();
    if (!permission.granted) return;

    const result = await ImagePicker.launchCameraAsync({
      mediaTypes: ['images'],
      quality: 0.8,
      allowsEditing: false,
    });

    if (!result.canceled && result.assets[0]) {
      const asset = result.assets[0];
      onClose();
      onPick({
        uri: asset.uri,
        filename: asset.fileName || `photo_${Date.now()}.jpg`,
        mimeType: asset.mimeType || 'image/jpeg',
        size: asset.fileSize,
        width: asset.width,
        height: asset.height,
        type: 'image',
      });
    }
  };

  const handleGallery = async () => {
    const permission = await ImagePicker.requestMediaLibraryPermissionsAsync();
    if (!permission.granted) return;

    const result = await ImagePicker.launchImageLibraryAsync({
      mediaTypes: ['images'],
      quality: 0.8,
      allowsEditing: false,
    });

    if (!result.canceled && result.assets[0]) {
      const asset = result.assets[0];
      onClose();
      onPick({
        uri: asset.uri,
        filename: asset.fileName || `image_${Date.now()}.jpg`,
        mimeType: asset.mimeType || 'image/jpeg',
        size: asset.fileSize,
        width: asset.width,
        height: asset.height,
        type: 'image',
      });
    }
  };

  const handleFile = async () => {
    const result = await DocumentPicker.getDocumentAsync({
      type: [
        'application/pdf',
        'application/msword',
        'application/vnd.openxmlformats-officedocument.wordprocessingml.document',
        'application/vnd.ms-excel',
        'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet',
        'application/vnd.ms-powerpoint',
        'application/vnd.openxmlformats-officedocument.presentationml.presentation',
        'text/plain',
        'application/zip',
      ],
      copyToCacheDirectory: true,
    });

    if (!result.canceled && result.assets[0]) {
      const asset = result.assets[0];
      onClose();
      onPick({
        uri: asset.uri,
        filename: asset.name,
        mimeType: asset.mimeType || 'application/octet-stream',
        size: asset.size,
        type: 'file',
      });
    }
  };

  const handlePress = (key: string) => {
    switch (key) {
      case 'camera':
        handleCamera();
        break;
      case 'gallery':
        handleGallery();
        break;
      case 'file':
        handleFile();
        break;
    }
  };

  return (
    <Modal
      visible={visible}
      transparent
      animationType="slide"
      onRequestClose={onClose}
    >
      <Pressable style={styles.overlay} onPress={onClose}>
        <View style={styles.sheet}>
          <View style={styles.handle} />
          <Text style={styles.title}>Kirim Media</Text>
          <View style={styles.options}>
            {OPTIONS.map((opt) => (
              <Pressable
                key={opt.key}
                style={({ pressed }) => [styles.option, pressed && styles.pressed]}
                onPress={() => handlePress(opt.key)}
              >
                <View style={styles.iconCircle}>
                  <Text style={styles.icon}>{opt.icon}</Text>
                </View>
                <Text style={styles.label}>{opt.label}</Text>
              </Pressable>
            ))}
          </View>
          <Pressable style={styles.cancelButton} onPress={onClose}>
            <Text style={styles.cancelText}>Batal</Text>
          </Pressable>
        </View>
      </Pressable>
    </Modal>
  );
}

export type { PickedMedia };

const styles = StyleSheet.create({
  overlay: {
    flex: 1,
    justifyContent: 'flex-end',
    backgroundColor: 'rgba(0,0,0,0.5)',
  },
  sheet: {
    backgroundColor: colors.surface,
    borderTopLeftRadius: 20,
    borderTopRightRadius: 20,
    paddingBottom: spacing.xxxl,
    paddingHorizontal: spacing.xl,
  },
  handle: {
    width: 40,
    height: 4,
    borderRadius: 2,
    backgroundColor: colors.border,
    alignSelf: 'center',
    marginTop: spacing.sm,
    marginBottom: spacing.lg,
  },
  title: {
    fontFamily: fontFamily.uiSemiBold,
    fontSize: fontSize.lg,
    color: colors.textPrimary,
    marginBottom: spacing.xl,
    textAlign: 'center',
  },
  options: {
    flexDirection: 'row',
    justifyContent: 'space-around',
    marginBottom: spacing.xxl,
  },
  option: {
    alignItems: 'center',
    gap: spacing.sm,
  },
  pressed: {
    opacity: 0.7,
  },
  iconCircle: {
    width: 56,
    height: 56,
    borderRadius: 28,
    backgroundColor: colors.surface2,
    justifyContent: 'center',
    alignItems: 'center',
  },
  icon: {
    fontSize: 24,
  },
  label: {
    fontFamily: fontFamily.ui,
    fontSize: fontSize.sm,
    color: colors.textPrimary,
  },
  cancelButton: {
    paddingVertical: spacing.md,
    alignItems: 'center',
  },
  cancelText: {
    fontFamily: fontFamily.ui,
    fontSize: fontSize.md,
    color: colors.textMuted,
  },
});
