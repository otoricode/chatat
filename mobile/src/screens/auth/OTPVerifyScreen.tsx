// OTP verification screen — 6-digit code input
import React, { useState, useRef, useCallback } from 'react';
import { View, Text, StyleSheet, TextInput, Pressable } from 'react-native';
import { SafeAreaView } from 'react-native-safe-area-context';
import type { NativeStackScreenProps } from '@react-navigation/native-stack';
import type { AuthStackParamList } from '@/navigation/types';
import { useTranslation } from 'react-i18next';
import { colors, fontSize, spacing, fontFamily } from '@/theme';
import { OTP_LENGTH, OTP_EXPIRY_SECONDS } from '@/lib/constants';

type Props = NativeStackScreenProps<AuthStackParamList, 'OTPVerify'>;

export function OTPVerifyScreen({ route, navigation }: Props) {
  const { t } = useTranslation();
  const { phone } = route.params;
  const [otp, setOtp] = useState('');
  const [isLoading, setIsLoading] = useState(false);
  const [countdown, setCountdown] = useState(OTP_EXPIRY_SECONDS);
  const inputRef = useRef<TextInput>(null);

  const canResend = countdown <= 0;

  const handleOtpChange = useCallback(
    (value: string) => {
      const cleaned = value.replace(/\D/g, '').slice(0, OTP_LENGTH);
      setOtp(cleaned);

      if (cleaned.length === OTP_LENGTH) {
        handleVerify(cleaned);
      }
    },
    [],
    // handleVerify is stable and does not need to be a dependency
  );

  const handleVerify = async (_code: string) => {
    setIsLoading(true);
    try {
      // TODO: Call auth API to verify OTP
      // On success, navigate to ProfileSetup (new user) or Main
      navigation.navigate('ProfileSetup');
    } finally {
      setIsLoading(false);
    }
  };

  const handleResend = () => {
    if (!canResend) return;
    setCountdown(OTP_EXPIRY_SECONDS);
    // TODO: Call API to resend OTP
  };

  // Countdown timer
  React.useEffect(() => {
    if (countdown <= 0) return;
    const timer = setInterval(() => {
      setCountdown((prev) => prev - 1);
    }, 1000);
    return () => clearInterval(timer);
  }, [countdown]);

  const maskedPhone = phone.replace(/(\+\d{2})(\d{3})(\d+)(\d{4})/, '$1 $2 **** $4');

  return (
    <SafeAreaView style={styles.container}>
      <View style={styles.content}>
        <Text style={styles.title}>{t('auth.enterOTP')}</Text>
        <Text style={styles.description}>
          Masukkan kode {OTP_LENGTH} digit yang dikirim ke {maskedPhone}
        </Text>

        <Pressable onPress={() => inputRef.current?.focus()} style={styles.otpContainer}>
          {Array.from({ length: OTP_LENGTH }).map((_, i) => (
            <View
              key={i}
              style={[styles.otpBox, i < otp.length && styles.otpBoxFilled]}
            >
              <Text style={styles.otpDigit}>{otp[i] ?? ''}</Text>
            </View>
          ))}
        </Pressable>

        <TextInput
          ref={inputRef}
          style={styles.hiddenInput}
          value={otp}
          onChangeText={handleOtpChange}
          keyboardType="number-pad"
          maxLength={OTP_LENGTH}
          autoFocus
        />

        {isLoading && <Text style={styles.loadingText}>Memverifikasi...</Text>}

        <View style={styles.resendRow}>
          {canResend ? (
            <Pressable onPress={handleResend}>
              <Text style={styles.resendActive}>Kirim Ulang</Text>
            </Pressable>
          ) : (
            <Text style={styles.resendDisabled}>
              Kirim ulang dalam {countdown} detik
            </Text>
          )}
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
  otpContainer: {
    flexDirection: 'row',
    justifyContent: 'center',
    gap: spacing.sm,
    marginBottom: spacing.xxl,
  },
  otpBox: {
    width: 48,
    height: 56,
    borderRadius: 12,
    backgroundColor: colors.surface,
    borderWidth: 1,
    borderColor: colors.border,
    justifyContent: 'center',
    alignItems: 'center',
  },
  otpBoxFilled: {
    borderColor: colors.green,
  },
  otpDigit: {
    fontFamily: fontFamily.uiBold,
    fontSize: fontSize.xxl,
    color: colors.textPrimary,
  },
  hiddenInput: {
    position: 'absolute',
    opacity: 0,
    height: 0,
    width: 0,
  },
  loadingText: {
    fontFamily: fontFamily.ui,
    fontSize: fontSize.sm,
    color: colors.green,
    textAlign: 'center',
    marginBottom: spacing.md,
  },
  resendRow: {
    alignItems: 'center',
  },
  resendActive: {
    fontFamily: fontFamily.uiSemiBold,
    fontSize: fontSize.sm,
    color: colors.green,
  },
  resendDisabled: {
    fontFamily: fontFamily.ui,
    fontSize: fontSize.sm,
    color: colors.textMuted,
  },
});
