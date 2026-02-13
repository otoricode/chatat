// Chat Stack Navigator — ChatList → Chat → ChatInfo, Topics, etc.
import React from 'react';
import { createNativeStackNavigator } from '@react-navigation/native-stack';
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
import { colors } from '@/theme';

const Stack = createNativeStackNavigator<ChatStackParamList>();

export function ChatStackNavigator() {
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
        options={{ title: 'Kontak' }}
      />
      <Stack.Screen
        name="CreateGroup"
        component={CreateGroupScreen}
        options={{ title: 'Buat Grup' }}
      />
      <Stack.Screen
        name="TopicList"
        component={TopicListScreen}
        options={{ title: 'Topik' }}
      />
      <Stack.Screen
        name="CreateTopic"
        component={CreateTopicScreen}
        options={{ title: 'Buat Topik' }}
      />
      <Stack.Screen
        name="Topic"
        component={TopicScreen}
        options={{ title: '' }}
      />
      <Stack.Screen
        name="TopicInfo"
        component={TopicInfoScreen}
        options={{ title: 'Info Topik' }}
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
        options={{ title: 'Dokumen' }}
      />
    </Stack.Navigator>
  );
}
