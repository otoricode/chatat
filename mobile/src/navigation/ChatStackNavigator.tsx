// Chat Stack Navigator — ChatList → Chat → ChatInfo, Topics, etc.
import React from 'react';
import { createNativeStackNavigator } from '@react-navigation/native-stack';
import { useTranslation } from 'react-i18next';
import type { ChatStackParamList } from './types';
import { ChatListScreen } from '@/screens/chat/ChatListScreen';
import { ChatScreen } from '@/screens/chat/ChatScreen';
import { ChatInfoScreen } from '@/screens/chat/ChatInfoScreen';
import { ContactListScreen } from '@/screens/contact/ContactListScreen';
import { CreateGroupScreen } from '@/screens/group/CreateGroupScreen';
import { TopicListScreen } from '@/screens/topic/TopicListScreen';
import { CreateTopicScreen } from '@/screens/topic/CreateTopicScreen';
import { TopicScreen } from '@/screens/topic/TopicScreen';
import { TopicInfoScreen } from '@/screens/topic/TopicInfoScreen';
import { ImageViewerScreen } from '@/screens/chat/ImageViewerScreen';
import { DocumentEditorScreen } from '@/screens/document/DocumentEditorScreen';
import { SearchScreen } from '@/screens/search/SearchScreen';
import { BackupScreen } from '@/screens/settings/BackupScreen';
import { SettingsScreen } from '@/screens/settings/SettingsScreen';
import { EditProfileScreen } from '@/screens/settings/EditProfileScreen';
import { LanguageScreen } from '@/screens/settings/LanguageScreen';
import { NotificationSettingsScreen } from '@/screens/settings/NotificationSettingsScreen';
import { StorageScreen } from '@/screens/settings/StorageScreen';
import { AboutScreen } from '@/screens/settings/AboutScreen';
import { PrivacySettingsScreen } from '@/screens/settings/PrivacySettingsScreen';
import { colors } from '@/theme';

const Stack = createNativeStackNavigator<ChatStackParamList>();

export function ChatStackNavigator() {
  const { t } = useTranslation();

  return (
    <Stack.Navigator
      screenOptions={{
        headerStyle: { backgroundColor: colors.headerBackground },
        headerTintColor: colors.textPrimary,
        headerTitleStyle: { fontFamily: 'PlusJakartaSans-SemiBold' },
        contentStyle: { backgroundColor: colors.background },
        animation: 'slide_from_right',
      }}
    >
      <Stack.Screen
        name="ChatList"
        component={ChatListScreen}
        options={{ headerShown: false }}
      />
      <Stack.Screen
        name="Chat"
        component={ChatScreen}
        options={{ headerShown: false }}
      />
      <Stack.Screen
        name="ChatInfo"
        component={ChatInfoScreen}
        options={{ title: '' }}
      />
      <Stack.Screen
        name="ContactList"
        component={ContactListScreen}
        options={{ title: t('contact.contacts') }}
      />
      <Stack.Screen
        name="CreateGroup"
        component={CreateGroupScreen}
        options={{ title: t('group.createGroup') }}
      />
      <Stack.Screen
        name="TopicList"
        component={TopicListScreen}
        options={{ title: t('topic.topics') }}
      />
      <Stack.Screen
        name="CreateTopic"
        component={CreateTopicScreen}
        options={{ title: t('topic.createTopic') }}
      />
      <Stack.Screen
        name="Topic"
        component={TopicScreen}
        options={{ title: '' }}
      />
      <Stack.Screen
        name="TopicInfo"
        component={TopicInfoScreen}
        options={{ title: t('topic.topicInfo') }}
      />
      <Stack.Screen
        name="ImageViewer"
        component={ImageViewerScreen}
        options={{
          title: '',
          animation: 'fade',
          headerStyle: { backgroundColor: '#000' },
        }}
      />
      <Stack.Screen
        name="DocumentEditor"
        component={DocumentEditorScreen}
        options={{ title: t('document.title') }}
      />
      <Stack.Screen
        name="Search"
        component={SearchScreen}
        options={{ title: t('search.title'), animation: 'fade' }}
      />
      <Stack.Screen
        name="Settings"
        component={SettingsScreen}
        options={{ title: t('settings.title') }}
      />
      <Stack.Screen
        name="EditProfile"
        component={EditProfileScreen}
        options={{ title: t('settings.editProfile') }}
      />
      <Stack.Screen
        name="Language"
        component={LanguageScreen}
        options={{ title: t('settings.language') }}
      />
      <Stack.Screen
        name="NotificationSettings"
        component={NotificationSettingsScreen}
        options={{ title: t('settings.notifications') }}
      />
      <Stack.Screen
        name="Storage"
        component={StorageScreen}
        options={{ title: t('settings.storage') }}
      />
      <Stack.Screen
        name="About"
        component={AboutScreen}
        options={{ title: t('settings.about') }}
      />
      <Stack.Screen
        name="PrivacySettings"
        component={PrivacySettingsScreen}
        options={{ title: t('privacy.title') }}
      />
      <Stack.Screen
        name="Backup"
        component={BackupScreen}
        options={{ title: t('backup.title') }}
      />
    </Stack.Navigator>
  );
}
