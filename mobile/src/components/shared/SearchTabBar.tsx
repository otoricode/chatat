// SearchTabBar — horizontal tab selector for search
import React from 'react';
import { View, Text, Pressable, ScrollView, StyleSheet } from 'react-native';
import { colors, fontFamily, fontSize, spacing } from '@/theme';
import type { SearchTab } from '@/hooks/useSearch';

type Tab = {
  key: SearchTab;
  label: string;
};

const TABS: Tab[] = [
  { key: 'all', label: 'Semua' },
  { key: 'messages', label: 'Pesan' },
  { key: 'documents', label: 'Dokumen' },
  { key: 'contacts', label: 'Kontak' },
  { key: 'entities', label: 'Entity' },
];

type SearchTabBarProps = {
  activeTab: SearchTab;
  onTabChange: (tab: SearchTab) => void;
};

export function SearchTabBar({ activeTab, onTabChange }: SearchTabBarProps) {
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
                {tab.label}
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
