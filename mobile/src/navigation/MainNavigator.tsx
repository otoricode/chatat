// Main Navigator — Bottom Tabs (Chat + Dokumen)
// Based on spesifikasi-chatat.md section 7.1
import React, { useCallback } from 'react';
import { StyleSheet, View, Text } from 'react-native';
import { createBottomTabNavigator } from '@react-navigation/bottom-tabs';
import { useNavigation, CommonActions } from '@react-navigation/native';
import type { MainTabParamList } from './types';
import { ChatStackNavigator } from './ChatStackNavigator';
import { DocumentStackNavigator } from './DocumentStackNavigator';
import { RealtimeProvider } from '@/components/shared/RealtimeProvider';
import { InAppNotification } from '@/components/shared/InAppNotification';
import { useNotifications } from '@/hooks/useNotifications';
import { useNotificationStore } from '@/stores/notificationStore';
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
  useNotifications();

  const { visible, title, body, data, hide } = useNotificationStore();
  const navigation = useNavigation();

  const handleNotificationPress = useCallback(() => {
    if (data.type === 'message' && data.chatId) {
      navigation.dispatch(
        CommonActions.navigate('ChatTab', {
          screen: 'Chat',
          params: { chatId: data.chatId, chatType: 'personal' },
        }),
      );
    } else if (data.type === 'group_message' && data.chatId) {
      navigation.dispatch(
        CommonActions.navigate('ChatTab', {
          screen: 'Chat',
          params: { chatId: data.chatId, chatType: 'group' },
        }),
      );
    } else if (data.type === 'topic_message' && data.topicId) {
      navigation.dispatch(
        CommonActions.navigate('ChatTab', {
          screen: 'Topic',
          params: { topicId: data.topicId },
        }),
      );
    } else if ((data.type === 'signature_request' || data.type === 'document_locked') && data.documentId) {
      navigation.dispatch(
        CommonActions.navigate('DocumentTab', {
          screen: 'DocumentEditor',
          params: { documentId: data.documentId },
        }),
      );
    } else if (data.type === 'group_invite' && data.chatId) {
      navigation.dispatch(
        CommonActions.navigate('ChatTab', {
          screen: 'Chat',
          params: { chatId: data.chatId, chatType: 'group' },
        }),
      );
    }
  }, [data, navigation]);

  return (
    <RealtimeProvider>
      <InAppNotification
        visible={visible}
        title={title}
        body={body}
        onPress={handleNotificationPress}
        onDismiss={hide}
      />
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
