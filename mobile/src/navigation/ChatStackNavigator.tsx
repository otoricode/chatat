// Chat Stack Navigator — ChatList → Chat → ChatInfo, Topics, etc.
import React from 'react';
import { createNativeStackNavigator } from '@react-navigation/native-stack';
import { useTranslation } from 'react-i18next';
import type { ChatStackParamList } from './types';
import { createLazyScreen } from '@/components/shared/LazyScreen';

// Eagerly loaded: critical path screens
import { ChatListScreen } from '@/screens/chat/ChatListScreen';
import { ChatScreen } from '@/screens/chat/ChatScreen';

// Lazily loaded: secondary screens
const ChatInfoScreen = createLazyScreen(() => import('@/screens/chat/ChatInfoScreen'), 'ChatInfoScreen');
const ContactListScreen = createLazyScreen(() => import('@/screens/contact/ContactListScreen'), 'ContactListScreen');
const CreateGroupScreen = createLazyScreen(() => import('@/screens/group/CreateGroupScreen'), 'CreateGroupScreen');
const TopicListScreen = createLazyScreen(() => import('@/screens/topic/TopicListScreen'), 'TopicListScreen');
const CreateTopicScreen = createLazyScreen(() => import('@/screens/topic/CreateTopicScreen'), 'CreateTopicScreen');
const TopicScreen = createLazyScreen(() => import('@/screens/topic/TopicScreen'), 'TopicScreen');
const TopicInfoScreen = createLazyScreen(() => import('@/screens/topic/TopicInfoScreen'), 'TopicInfoScreen');
const ImageViewerScreen = createLazyScreen(() => import('@/screens/chat/ImageViewerScreen'), 'ImageViewerScreen');
const DocumentEditorScreen = createLazyScreen(() => import('@/screens/document/DocumentEditorScreen'), 'DocumentEditorScreen');
const SearchScreen = createLazyScreen(() => import('@/screens/search/SearchScreen'), 'SearchScreen');
const BackupScreen = createLazyScreen(() => import('@/screens/settings/BackupScreen'), 'BackupScreen');
const SettingsScreen = createLazyScreen(() => import('@/screens/settings/SettingsScreen'), 'SettingsScreen');
const EditProfileScreen = createLazyScreen(() => import('@/screens/settings/EditProfileScreen'), 'EditProfileScreen');
const LanguageScreen = createLazyScreen(() => import('@/screens/settings/LanguageScreen'), 'LanguageScreen');
const NotificationSettingsScreen = createLazyScreen(() => import('@/screens/settings/NotificationSettingsScreen'), 'NotificationSettingsScreen');
const StorageScreen = createLazyScreen(() => import('@/screens/settings/StorageScreen'), 'StorageScreen');
const AboutScreen = createLazyScreen(() => import('@/screens/settings/AboutScreen'), 'AboutScreen');
const PrivacySettingsScreen = createLazyScreen(() => import('@/screens/settings/PrivacySettingsScreen'), 'PrivacySettingsScreen');
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
