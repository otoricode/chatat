// SearchTabBar — horizontal tab selector for search
import React from 'react';
import { View, Text, Pressable, ScrollView, StyleSheet } from 'react-native';
import { useTranslation } from 'react-i18next';
import { colors, fontFamily, fontSize, spacing } from '@/theme';
import type { SearchTab } from '@/hooks/useSearch';

type Tab = {
  key: SearchTab;
  labelKey: string;
};

const TABS: Tab[] = [
  { key: 'all', labelKey: 'search.all' },
  { key: 'messages', labelKey: 'search.messages' },
  { key: 'documents', labelKey: 'search.documents' },
  { key: 'contacts', labelKey: 'search.contacts' },
  { key: 'entities', labelKey: 'entity.entities' },
];

type SearchTabBarProps = {
  activeTab: SearchTab;
  onTabChange: (tab: SearchTab) => void;
};

export function SearchTabBar({ activeTab, onTabChange }: SearchTabBarProps) {
  const { t } = useTranslation();
  return (
    <View style={styles.wrapper}>
      <ScrollView
        horizontal
        showsHorizontalScrollIndicator={false}
        contentContainerStyle={styles.container}
      >
        {TABS.map((tab) => {
          const isActive = activeTab === tab.key;
          return (
            <Pressable
              key={tab.key}
              onPress={() => onTabChange(tab.key)}
              style={[styles.tab, isActive && styles.tabActive]}
            >
              <Text style={[styles.label, isActive && styles.labelActive]}>
                {t(tab.labelKey)}
              </Text>
            </Pressable>
          );
        })}
      </ScrollView>
    </View>
  );
}

const styles = StyleSheet.create({
  wrapper: {
    borderBottomWidth: 1,
    borderBottomColor: colors.border,
  },
  container: {
    paddingHorizontal: spacing.lg,
    gap: spacing.xs,
  },
  tab: {
    paddingHorizontal: spacing.md,
    paddingVertical: spacing.sm,
    borderBottomWidth: 2,
    borderBottomColor: colors.transparent,
  },
  tabActive: {
    borderBottomColor: colors.green,
  },
  label: {
    fontFamily: fontFamily.uiMedium,
    fontSize: fontSize.sm,
    color: colors.textMuted,
  },
  labelActive: {
    color: colors.green,
  },
});
