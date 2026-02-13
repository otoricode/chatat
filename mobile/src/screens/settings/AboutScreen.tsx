// About Screen — app info, version, credits
import React from 'react';
import { View, Text, ScrollView, StyleSheet, Linking, TouchableOpacity } from 'react-native';
import { useTranslation } from 'react-i18next';
import { colors, fontFamily, fontSize, spacing } from '@/theme';

const APP_VERSION = '1.0.0';
const BUILD_NUMBER = '1';

export function AboutScreen() {
  const { t } = useTranslation();

  return (
    <ScrollView style={styles.container} contentContainerStyle={styles.content}>
      {/* App Identity */}
      <View style={styles.appHeader}>
        <Text style={styles.appIcon}>💬</Text>
        <Text style={styles.appName}>Chatat</Text>
        <Text style={styles.appTagline}>{t('settings.aboutTagline')}</Text>
      </View>

      {/* Version Info */}
      <View style={styles.card}>
        <View style={styles.infoRow}>
          <Text style={styles.infoLabel}>{t('settings.version')}</Text>
          <Text style={styles.infoValue}>{APP_VERSION}</Text>
        </View>
        <View style={styles.divider} />
        <View style={styles.infoRow}>
          <Text style={styles.infoLabel}>{t('settings.build')}</Text>
          <Text style={styles.infoValue}>{BUILD_NUMBER}</Text>
        </View>
      </View>

      {/* Links */}
      <View style={styles.card}>
        <TouchableOpacity
          style={styles.linkRow}
          onPress={() => Linking.openURL('https://otoritech.com')}
          activeOpacity={0.6}
        >
          <Text style={styles.linkText}>{t('settings.website')}</Text>
          <Text style={styles.linkChevron}>›</Text>
        </TouchableOpacity>
        <View style={styles.divider} />
        <TouchableOpacity
          style={styles.linkRow}
          onPress={() => Linking.openURL('https://otoritech.com/privacy')}
          activeOpacity={0.6}
        >
          <Text style={styles.linkText}>{t('settings.privacyPolicy')}</Text>
          <Text style={styles.linkChevron}>›</Text>
        </TouchableOpacity>
        <View style={styles.divider} />
        <TouchableOpacity
          style={styles.linkRow}
          onPress={() => Linking.openURL('https://otoritech.com/terms')}
          activeOpacity={0.6}
        >
          <Text style={styles.linkText}>{t('settings.termsOfService')}</Text>
          <Text style={styles.linkChevron}>›</Text>
        </TouchableOpacity>
      </View>

      {/* Copyright */}
      <Text style={styles.copyright}>
        {'\u00A9'} 2025 Otoritech. All rights reserved.
      </Text>
    </ScrollView>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: colors.background,
  },
  content: {
    padding: spacing.lg,
    paddingBottom: spacing.xxxl,
    alignItems: 'center',
  },
  appHeader: {
    alignItems: 'center',
    marginBottom: spacing.xl,
    paddingVertical: spacing.xl,
  },
  appIcon: {
    fontSize: 64,
    marginBottom: spacing.md,
  },
  appName: {
    fontFamily: fontFamily.uiBold,
    fontSize: fontSize.h1,
    color: colors.green,
  },
  appTagline: {
    fontFamily: fontFamily.ui,
    fontSize: fontSize.sm,
    color: colors.textMuted,
    marginTop: spacing.xs,
  },
  card: {
    width: '100%',
    backgroundColor: colors.surface,
    borderRadius: 12,
    marginBottom: spacing.lg,
    overflow: 'hidden',
  },
  infoRow: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    paddingHorizontal: spacing.lg,
    paddingVertical: spacing.md,
  },
  infoLabel: {
    fontFamily: fontFamily.ui,
    fontSize: fontSize.md,
    color: colors.textMuted,
  },
  infoValue: {
    fontFamily: fontFamily.uiMedium,
    fontSize: fontSize.md,
    color: colors.textPrimary,
  },
  divider: {
    height: StyleSheet.hairlineWidth,
    backgroundColor: colors.border,
    marginLeft: spacing.lg,
  },
  linkRow: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    paddingHorizontal: spacing.lg,
    paddingVertical: spacing.md,
  },
  linkText: {
    fontFamily: fontFamily.ui,
    fontSize: fontSize.md,
    color: colors.textPrimary,
  },
  linkChevron: {
    fontFamily: fontFamily.ui,
    fontSize: fontSize.xl,
    color: colors.textMuted,
  },
  copyright: {
    fontFamily: fontFamily.ui,
    fontSize: fontSize.xs,
    color: colors.textMuted,
    textAlign: 'center',
    marginTop: spacing.lg,
  },
});
