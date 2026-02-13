// Phone input screen — first step of auth flow
import React, { useState } from 'react';
import { View, Text, StyleSheet, TextInput, Pressable } from 'react-native';
import { SafeAreaView } from 'react-native-safe-area-context';
import type { NativeStackScreenProps } from '@react-navigation/native-stack';
import type { AuthStackParamList } from '@/navigation/types';
import { useTranslation } from 'react-i18next';
import { colors, fontSize, spacing, fontFamily } from '@/theme';

type Props = NativeStackScreenProps<AuthStackParamList, 'PhoneInput'>;

export function PhoneInputScreen({ navigation }: Props) {
  const { t } = useTranslation();
  const [phone, setPhone] = useState('');
  const [countryCode] = useState('+62');

  const isValid = phone.length >= 9;

  const handleSMSOTP = () => {
    if (!isValid) return;
    const fullPhone = `${countryCode}${phone.replace(/^0/, '')}`;
    navigation.navigate('OTPVerify', { phone: fullPhone, method: 'sms' });
  };

  const handleReverseOTP = () => {
    if (!isValid) return;
    // TODO: Call API to init reverse OTP, then navigate
    navigation.navigate('OTPVerify', {
      phone: `${countryCode}${phone.replace(/^0/, '')}`,
      method: 'reverse',
    });
  };

  return (
    <SafeAreaView style={styles.container}>
      <View style={styles.content}>
        <Text style={styles.logo}>Chatat</Text>
        <Text style={styles.subtitle}>Chat + Collaboration</Text>

        <View style={styles.inputSection}>
          <Text style={styles.label}>{t('auth.enterPhone')}</Text>
          <View style={styles.phoneRow}>
            <View style={styles.countryCode}>
              <Text style={styles.countryCodeText}>{countryCode}</Text>
            </View>
            <TextInput
              style={styles.phoneInput}
              value={phone}
              onChangeText={setPhone}
              placeholder="812 3456 7890"
              placeholderTextColor={colors.textMuted}
              keyboardType="phone-pad"
              maxLength={15}
              autoFocus
            />
          </View>
        </View>

        <Pressable
          style={[styles.button, styles.buttonPrimary, !isValid && styles.buttonDisabled]}
          onPress={handleSMSOTP}
          disabled={!isValid}
        >
          <Text style={styles.buttonText}>{t('auth.sendOTP')}</Text>
        </Pressable>

        <Pressable
          style={[styles.button, styles.buttonSecondary, !isValid && styles.buttonDisabled]}
          onPress={handleReverseOTP}
          disabled={!isValid}
        >
          <Text style={[styles.buttonText, styles.buttonSecondaryText]}>
            {t('common.next')}
          </Text>
        </Pressable>
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
  logo: {
    fontFamily: fontFamily.uiBold,
    fontSize: fontSize.h1,
    color: colors.green,
    textAlign: 'center',
    marginBottom: spacing.xs,
  },
  subtitle: {
    fontFamily: fontFamily.ui,
    fontSize: fontSize.md,
    color: colors.textMuted,
    textAlign: 'center',
    marginBottom: spacing.xxxl,
  },
  inputSection: {
    marginBottom: spacing.xxl,
  },
  label: {
    fontFamily: fontFamily.uiMedium,
    fontSize: fontSize.sm,
    color: colors.textPrimary,
    marginBottom: spacing.sm,
  },
  phoneRow: {
    flexDirection: 'row',
    gap: spacing.sm,
  },
  countryCode: {
    backgroundColor: colors.surface,
    borderRadius: 12,
    paddingHorizontal: spacing.lg,
    justifyContent: 'center',
    borderWidth: 1,
    borderColor: colors.border,
  },
  countryCodeText: {
    fontFamily: fontFamily.uiMedium,
    fontSize: fontSize.lg,
    color: colors.textPrimary,
  },
  phoneInput: {
    flex: 1,
    backgroundColor: colors.surface,
    borderRadius: 12,
    paddingHorizontal: spacing.lg,
    paddingVertical: spacing.md,
    fontFamily: fontFamily.ui,
    fontSize: fontSize.lg,
    color: colors.textPrimary,
    borderWidth: 1,
    borderColor: colors.border,
  },
  button: {
    borderRadius: 12,
    paddingVertical: spacing.lg,
    alignItems: 'center',
    marginBottom: spacing.md,
  },
  buttonPrimary: {
    backgroundColor: colors.green,
  },
  buttonSecondary: {
    backgroundColor: colors.transparent,
    borderWidth: 1,
    borderColor: colors.green,
  },
  buttonDisabled: {
    opacity: 0.4,
  },
  buttonText: {
    fontFamily: fontFamily.uiSemiBold,
    fontSize: fontSize.md,
    color: colors.background,
  },
  buttonSecondaryText: {
    color: colors.green,
  },
});
