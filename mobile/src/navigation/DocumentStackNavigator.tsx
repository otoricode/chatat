// Document Stack Navigator — DocumentList → Editor/Viewer
import React from 'react';
import { createNativeStackNavigator } from '@react-navigation/native-stack';
import type { DocumentStackParamList } from './types';
import { DocumentListScreen } from '@/screens/document/DocumentListScreen';
import { DocumentEditorScreen } from '@/screens/document/DocumentEditorScreen';
import { DocumentViewerScreen } from '@/screens/document/DocumentViewerScreen';
import { EntityListScreen } from '@/screens/entity/EntityListScreen';
import { EntityDetailScreen } from '@/screens/entity/EntityDetailScreen';
import { colors } from '@/theme';

const Stack = createNativeStackNavigator<DocumentStackParamList>();

export function DocumentStackNavigator() {
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
        name="DocumentList"
        component={DocumentListScreen}
        options={{ headerShown: false }}
      />
      <Stack.Screen
        name="DocumentEditor"
        component={DocumentEditorScreen}
        options={{ title: 'Dokumen' }}
      />
      <Stack.Screen
        name="DocumentViewer"
        component={DocumentViewerScreen}
        options={{ title: 'Dokumen' }}
      />
      <Stack.Screen
        name="EntityList"
        component={EntityListScreen}
        options={{ headerShown: false }}
      />
      <Stack.Screen
        name="EntityDetail"
        component={EntityDetailScreen}
        options={{ title: 'Entity' }}
      />
    </Stack.Navigator>
  );
}
