// Network offline banner — shows when device is offline
import React, { useEffect, useRef } from 'react';
import { Animated, StyleSheet, Text, View } from 'react-native';
import { useTranslation } from 'react-i18next';
import { useNetworkStore } from '@/stores/networkStore';
import { colors } from '@/theme';

export const NetworkBanner: React.FC = () => {
  const { t } = useTranslation();
  const isConnected = useNetworkStore((s) => s.isConnected);
  const slideAnim = useRef(new Animated.Value(-50)).current;

  useEffect(() => {
    Animated.timing(slideAnim, {
      toValue: isConnected ? -50 : 0,
      duration: 300,
      useNativeDriver: true,
    }).start();
  }, [isConnected, slideAnim]);

  if (isConnected) return null;

  return (
    <Animated.View
      style={[
        styles.banner,
        { transform: [{ translateY: slideAnim }] },
      ]}
    >
      <View style={styles.content}>
        <Text style={styles.icon}>📡</Text>
        <Text style={styles.text}>{t('network.offline')}</Text>
      </View>
    </Animated.View>
  );
};

const styles = StyleSheet.create({
  banner: {
    backgroundColor: colors.surface2,
    paddingVertical: 8,
    paddingHorizontal: 16,
    borderBottomWidth: 1,
    borderBottomColor: colors.border,
  },
  content: {
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'center',
    gap: 8,
  },
  icon: {
    fontSize: 14,
  },
  text: {
    color: colors.yellow,
    fontSize: 13,
    fontWeight: '500',
  },
});
