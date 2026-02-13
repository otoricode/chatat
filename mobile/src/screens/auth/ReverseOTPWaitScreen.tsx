// Reverse OTP wait screen — show WA number + code to send
import React, { useState } from 'react';
import { View, Text, StyleSheet, Pressable, Linking } from 'react-native';
import { SafeAreaView } from 'react-native-safe-area-context';
import type { NativeStackScreenProps } from '@react-navigation/native-stack';
import type { AuthStackParamList } from '@/navigation/types';
import { colors, fontSize, spacing, fontFamily } from '@/theme';
import { useTranslation } from 'react-i18next';

type Props = NativeStackScreenProps<AuthStackParamList, 'ReverseOTPWait'>;

const REVERSE_OTP_TIMEOUT = 300; // 5 minutes

export function ReverseOTPWaitScreen({ route, navigation }: Props) {
  const { waNumber, code } = route.params;
  const { t } = useTranslation();
  const [countdown, setCountdown] = useState(REVERSE_OTP_TIMEOUT);

  React.useEffect(() => {
    if (countdown <= 0) return;
    const timer = setInterval(() => {
      setCountdown((prev) => prev - 1);
    }, 1000);
    return () => clearInterval(timer);
  }, [countdown]);

  // TODO: Poll API for verification status
  React.useEffect(() => {
    const interval = setInterval(() => {
      // Check if reverse OTP has been verified
      // If verified, navigate to ProfileSetup or Main
    }, 3000);
    return () => clearInterval(interval);
  }, [navigation]);

  const handleOpenWhatsApp = () => {
    const url = `whatsapp://send?phone=${waNumber.replace('+', '')}&text=${code}`;
    Linking.openURL(url).catch(() => {
      // WhatsApp not installed
    });
  };

  const formatTime = (seconds: number) => {
    const m = Math.floor(seconds / 60);
    const s = seconds % 60;
    return `${m}:${s.toString().padStart(2, '0')}`;
  };

  return (
    <SafeAreaView style={styles.container}>
      <View style={styles.content}>
        <Text style={styles.title}>{t('auth.waVerification')}</Text>
        <Text style={styles.description}>
          {t('auth.waVerificationDesc')}
        </Text>

        <View style={styles.infoCard}>
          <Text style={styles.cardLabel}>{t('auth.waNumber')}</Text>
          <Text style={styles.waNumber}>{waNumber}</Text>
        </View>

        <View style={styles.infoCard}>
          <Text style={styles.cardLabel}>{t('auth.verificationCode')}</Text>
          <Text style={styles.code}>{code}</Text>
        </View>

        <Pressable style={styles.waButton} onPress={handleOpenWhatsApp}>
          <Text style={styles.waButtonText}>{t('auth.openWhatsApp')}</Text>
        </Pressable>

        <View style={styles.waitingSection}>
          <Text style={styles.waitingText}>{t('auth.waitingVerification')}</Text>
          <Text style={styles.timerText}>
            {t('auth.validFor', { time: formatTime(countdown) })}
          </Text>
        </View>
      </View>
    </SafeAreaView>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: colors.background,
  },
  content: {
    flex: 1,
    paddingHorizontal: spacing.xxl,
    justifyContent: 'center',
  },
  title: {
    fontFamily: fontFamily.uiBold,
    fontSize: fontSize.xl,
    color: colors.textPrimary,
    textAlign: 'center',
    marginBottom: spacing.sm,
  },
  description: {
    fontFamily: fontFamily.ui,
    fontSize: fontSize.sm,
    color: colors.textMuted,
    textAlign: 'center',
    marginBottom: spacing.xxxl,
  },
  infoCard: {
    backgroundColor: colors.surface,
    borderRadius: 12,
    padding: spacing.lg,
    marginBottom: spacing.lg,
    alignItems: 'center',
    borderWidth: 1,
    borderColor: colors.border,
  },
  cardLabel: {
    fontFamily: fontFamily.ui,
    fontSize: fontSize.xs,
    color: colors.textMuted,
    marginBottom: spacing.xs,
  },
  waNumber: {
    fontFamily: fontFamily.uiBold,
    fontSize: fontSize.xl,
    color: colors.textPrimary,
  },
  code: {
    fontFamily: fontFamily.uiBold,
    fontSize: fontSize.h1,
    color: colors.green,
    letterSpacing: 4,
  },
  waButton: {
    backgroundColor: '#25D366',
    borderRadius: 12,
    paddingVertical: spacing.lg,
    alignItems: 'center',
    marginBottom: spacing.xxl,
  },
  waButtonText: {
    fontFamily: fontFamily.uiSemiBold,
    fontSize: fontSize.md,
    color: colors.white,
  },
  waitingSection: {
    alignItems: 'center',
  },
  waitingText: {
    fontFamily: fontFamily.ui,
    fontSize: fontSize.sm,
    color: colors.green,
    marginBottom: spacing.xs,
  },
  timerText: {
    fontFamily: fontFamily.ui,
    fontSize: fontSize.xs,
    color: colors.textMuted,
  },
});
