// Language Screen — select app language
import React, { useCallback } from 'react';
import { View, Text, StyleSheet, TouchableOpacity, FlatList } from 'react-native';
import { useTranslation } from 'react-i18next';
import { setLanguage, getCurrentLanguage } from '@/i18n';
import type { SupportedLanguage } from '@/i18n';
import { colors, fontFamily, fontSize, spacing } from '@/theme';

type LanguageItem = {
  code: SupportedLanguage;
  label: string;
  flag: string;
};

const LANGUAGES: LanguageItem[] = [
  { code: 'id', label: 'Bahasa Indonesia', flag: '🇮🇩' },
  { code: 'en', label: 'English', flag: '🇬🇧' },
  { code: 'ar', label: '\u0627\u0644\u0639\u0631\u0628\u064a\u0629', flag: '🇸🇦' },
];

export function LanguageScreen() {
  const { i18n } = useTranslation();
  const currentLang = getCurrentLanguage();

  const handleSelect = useCallback(
    async (code: SupportedLanguage) => {
      if (code === i18n.language) return;
      await setLanguage(code);
    },
    [i18n.language],
  );

  const renderItem = useCallback(
    ({ item }: { item: LanguageItem }) => {
      const isActive = currentLang === item.code;
      return (
        <TouchableOpacity
          style={[styles.row, isActive && styles.rowActive]}
          onPress={() => handleSelect(item.code)}
          activeOpacity={0.6}
        >
          <Text style={styles.flag}>{item.flag}</Text>
          <Text style={[styles.label, isActive && styles.labelActive]}>
            {item.label}
          </Text>
          {isActive && <Text style={styles.check}>✓</Text>}
        </TouchableOpacity>
      );
    },
    [currentLang, handleSelect],
  );

  return (
    <View style={styles.container}>
      <FlatList
        data={LANGUAGES}
        keyExtractor={(item) => item.code}
        renderItem={renderItem}
        contentContainerStyle={styles.list}
      />
    </View>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: colors.background,
  },
  list: {
    padding: spacing.lg,
  },
  row: {
    flexDirection: 'row',
    alignItems: 'center',
    backgroundColor: colors.surface,
    borderRadius: 12,
    padding: spacing.lg,
    marginBottom: spacing.sm,
  },
  rowActive: {
    borderWidth: 1,
    borderColor: colors.green,
  },
  flag: {
    fontSize: 24,
    marginRight: spacing.md,
  },
  label: {
    flex: 1,
    fontFamily: fontFamily.ui,
    fontSize: fontSize.md,
    color: colors.textPrimary,
  },
  labelActive: {
    fontFamily: fontFamily.uiSemiBold,
    color: colors.green,
  },
  check: {
    fontFamily: fontFamily.uiSemiBold,
    fontSize: fontSize.lg,
    color: colors.green,
  },
});
