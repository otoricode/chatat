// Tab bar for switching between Chat, Documents, and Topics tabs within a chat screen
import React from 'react';
import { View, Text, Pressable, StyleSheet } from 'react-native';
import { colors, fontSize, fontFamily, spacing } from '@/theme';

export type ChatTab = 'chat' | 'documents' | 'topics';

type TabItem = {
  key: ChatTab;
  label: string;
  icon: string;
};

type ChatTabBarProps = {
  tabs: TabItem[];
  activeTab: ChatTab;
  onTabChange: (tab: ChatTab) => void;
};

export function ChatTabBar({ tabs, activeTab, onTabChange }: ChatTabBarProps) {
  return (
    <View style={styles.container}>
      {tabs.map((tab) => {
        const isActive = tab.key === activeTab;
        return (
          <Pressable
            key={tab.key}
            style={[styles.tab, isActive && styles.activeTab]}
            onPress={() => onTabChange(tab.key)}
          >
            <Text style={styles.icon}>{tab.icon}</Text>
            <Text style={[styles.label, isActive && styles.activeLabel]}>
              {tab.label}
            </Text>
          </Pressable>
        );
      })}
    </View>
  );
}

const styles = StyleSheet.create({
  container: {
    flexDirection: 'row',
    backgroundColor: colors.surface,
    borderBottomWidth: 1,
    borderBottomColor: colors.border,
  },
  tab: {
    flex: 1,
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'center',
    paddingVertical: spacing.sm,
    gap: spacing.xs,
    borderBottomWidth: 2,
    borderBottomColor: 'transparent',
  },
  activeTab: {
    borderBottomColor: colors.green,
  },
  icon: {
    fontSize: fontSize.sm,
  },
  label: {
    fontFamily: fontFamily.ui,
    fontSize: fontSize.sm,
    color: colors.textMuted,
  },
  activeLabel: {
    fontFamily: fontFamily.uiSemiBold,
    color: colors.green,
  },
});
