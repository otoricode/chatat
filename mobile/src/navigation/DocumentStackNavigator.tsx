// Document Stack Navigator — DocumentList → Editor/Viewer
import React from 'react';
import { createNativeStackNavigator } from '@react-navigation/native-stack';
import { useTranslation } from 'react-i18next';
import type { DocumentStackParamList } from './types';
import { DocumentListScreen } from '@/screens/document/DocumentListScreen';
import { DocumentEditorScreen } from '@/screens/document/DocumentEditorScreen';
import { DocumentViewerScreen } from '@/screens/document/DocumentViewerScreen';
import { EntityListScreen } from '@/screens/entity/EntityListScreen';
import { EntityDetailScreen } from '@/screens/entity/EntityDetailScreen';
import { colors } from '@/theme';

const Stack = createNativeStackNavigator<DocumentStackParamList>();

export function DocumentStackNavigator() {
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
        name="DocumentList"
        component={DocumentListScreen}
        options={{ headerShown: false }}
      />
      <Stack.Screen
        name="DocumentEditor"
        component={DocumentEditorScreen}
        options={{ title: t('document.title') }}
      />
      <Stack.Screen
        name="DocumentViewer"
        component={DocumentViewerScreen}
        options={{ title: t('document.title') }}
      />
      <Stack.Screen
        name="EntityList"
        component={EntityListScreen}
        options={{ headerShown: false }}
      />
      <Stack.Screen
        name="EntityDetail"
        component={EntityDetailScreen}
        options={{ title: t('entity.entities') }}
      />
    </Stack.Navigator>
  );
}
