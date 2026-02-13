// Main Navigator — Bottom Tabs (Chat + Dokumen)
// Based on spesifikasi-chatat.md section 7.1
import React from 'react';
import { StyleSheet, View, Text } from 'react-native';
import { createBottomTabNavigator } from '@react-navigation/bottom-tabs';
import type { MainTabParamList } from './types';
import { ChatStackNavigator } from './ChatStackNavigator';
import { DocumentStackNavigator } from './DocumentStackNavigator';
import { RealtimeProvider } from '@/components/shared/RealtimeProvider';
import { colors, fontSize } from '@/theme';

const Tab = createBottomTabNavigator<MainTabParamList>();

function ChatTabIcon({ focused }: { focused: boolean }) {
  return (
    <View style={styles.iconContainer}>
      <Text style={[styles.icon, focused && styles.iconFocused]}>💬</Text>
    </View>
  );
}

function DocumentTabIcon({ focused }: { focused: boolean }) {
  return (
    <View style={styles.iconContainer}>
      <Text style={[styles.icon, focused && styles.iconFocused]}>📄</Text>
    </View>
  );
}

export function MainNavigator() {
  return (
    <RealtimeProvider>
    <Tab.Navigator
      screenOptions={{
        headerShown: false,
        tabBarStyle: styles.tabBar,
        tabBarActiveTintColor: colors.green,
        tabBarInactiveTintColor: colors.textMuted,
        tabBarLabelStyle: styles.tabLabel,
      }}
    >
      <Tab.Screen
        name="ChatTab"
        component={ChatStackNavigator}
        options={{
          tabBarLabel: 'Chat',
          tabBarIcon: ChatTabIcon,
        }}
      />
      <Tab.Screen
        name="DocumentTab"
        component={DocumentStackNavigator}
        options={{
          tabBarLabel: 'Dokumen',
          tabBarIcon: DocumentTabIcon,
        }}
      />
    </Tab.Navigator>
    </RealtimeProvider>
  );
}

const styles = StyleSheet.create({
  tabBar: {
    backgroundColor: colors.tabBarBackground,
    borderTopColor: colors.border,
    borderTopWidth: 0.5,
    height: 60,
    paddingBottom: 8,
    paddingTop: 8,
  },
  tabLabel: {
    fontFamily: 'PlusJakartaSans-Medium',
    fontSize: fontSize.xs,
  },
  iconContainer: {
    alignItems: 'center',
    justifyContent: 'center',
  },
  icon: {
    fontSize: 22,
    opacity: 0.6,
  },
  iconFocused: {
    opacity: 1,
  },
});
