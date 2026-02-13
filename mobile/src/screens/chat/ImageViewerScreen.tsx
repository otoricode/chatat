// ImageViewerScreen — full-screen image viewer with pinch-to-zoom
import React, { useCallback, useState } from 'react';
import {
  View,
  Text,
  Pressable,
  StyleSheet,
  Dimensions,
  ActivityIndicator,
  Alert,
  ScrollView,
  Share,
} from 'react-native';
import { Image } from 'expo-image';
import { SafeAreaView } from 'react-native-safe-area-context';
import type { NativeStackScreenProps } from '@react-navigation/native-stack';
import type { ChatStackParamList } from '@/navigation/types';
import { colors, spacing } from '@/theme';
import { useTranslation } from 'react-i18next';

type Props = NativeStackScreenProps<ChatStackParamList, 'ImageViewer'>;

const { width: SCREEN_WIDTH, height: SCREEN_HEIGHT } = Dimensions.get('window');

export function ImageViewerScreen({ route, navigation }: Props) {
  const { url, filename } = route.params;
  const { t } = useTranslation();
  const [loading, setLoading] = useState(true);

  React.useLayoutEffect(() => {
    navigation.setOptions({
      headerShown: true,
      headerStyle: { backgroundColor: '#000' },
      headerTintColor: colors.white,
      headerTitle: filename || t('media.image'),
      headerRight: () => (
        <Pressable onPress={handleShare} style={headerStyles.button}>
          <Text style={headerStyles.shareIcon}>{'\u{2B06}'}</Text>
        </Pressable>
      ),
    });
  }, [navigation, filename]);

  const handleShare = useCallback(async () => {
    try {
      await Share.share({
        url: url,
        message: filename || t('media.image'),
      });
    } catch {
      Alert.alert(t('common.failed'), t('media.shareFailed'));
    }
  }, [url, filename]);

  return (
    <SafeAreaView style={styles.container} edges={['bottom']}>
      <View style={styles.imageContainer}>
        <ScrollView
          style={styles.scrollView}
          contentContainerStyle={styles.scrollContent}
          maximumZoomScale={4}
          minimumZoomScale={1}
          showsHorizontalScrollIndicator={false}
          showsVerticalScrollIndicator={false}
          centerContent
        >
          <Image
            source={{ uri: url }}
            style={styles.image}
            contentFit="contain"
            cachePolicy="disk"
            transition={200}
            onLoadEnd={() => setLoading(false)}
          />
        </ScrollView>

        {loading && (
          <View style={styles.loadingOverlay}>
            <ActivityIndicator size="large" color={colors.white} />
          </View>
        )}
      </View>
    </SafeAreaView>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: '#000',
  },
  imageContainer: {
    flex: 1,
    justifyContent: 'center',
    alignItems: 'center',
  },
  scrollView: {
    flex: 1,
    width: SCREEN_WIDTH,
  },
  scrollContent: {
    flexGrow: 1,
    justifyContent: 'center',
    alignItems: 'center',
  },
  image: {
    width: SCREEN_WIDTH,
    height: SCREEN_HEIGHT * 0.8,
  },
  loadingOverlay: {
    ...StyleSheet.absoluteFillObject,
    justifyContent: 'center',
    alignItems: 'center',
    backgroundColor: 'rgba(0,0,0,0.5)',
    gap: spacing.sm,
  },
});

const headerStyles = StyleSheet.create({
  button: {
    padding: spacing.sm,
  },
  shareIcon: {
    fontSize: 20,
    color: colors.white,
  },
});
