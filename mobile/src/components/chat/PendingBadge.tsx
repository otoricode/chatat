// Badge showing count of pending (unsent) messages
import React from 'react';
import { StyleSheet, Text, View } from 'react-native';
import { useTranslation } from 'react-i18next';
import { colors } from '@/theme';

interface PendingBadgeProps {
  count: number;
}

export const PendingBadge: React.FC<PendingBadgeProps> = ({ count }) => {
  const { t } = useTranslation();

  if (count === 0) return null;

  return (
    <View style={styles.container}>
      <Text style={styles.text}>
        {t('network.pendingMessages', { count })}
      </Text>
    </View>
  );
};

const styles = StyleSheet.create({
  container: {
    backgroundColor: colors.yellow,
    borderRadius: 4,
    paddingHorizontal: 8,
    paddingVertical: 2,
    alignSelf: 'flex-start',
  },
  text: {
    color: colors.background,
    fontSize: 11,
    fontWeight: '600',
  },
});
