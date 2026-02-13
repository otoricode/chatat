// Document Stack Navigator — DocumentList → Editor/Viewer
import React from 'react';
import { createNativeStackNavigator } from '@react-navigation/native-stack';
import { useTranslation } from 'react-i18next';
import type { DocumentStackParamList } from './types';
import { createLazyScreen } from '@/components/shared/LazyScreen';

// Eagerly loaded: critical path screen
import { DocumentListScreen } from '@/screens/document/DocumentListScreen';

// Lazily loaded: secondary screens
const DocumentEditorScreen = createLazyScreen(() => import('@/screens/document/DocumentEditorScreen'), 'DocumentEditorScreen');
const DocumentViewerScreen = createLazyScreen(() => import('@/screens/document/DocumentViewerScreen'), 'DocumentViewerScreen');
const EntityListScreen = createLazyScreen(() => import('@/screens/entity/EntityListScreen'), 'EntityListScreen');
const EntityDetailScreen = createLazyScreen(() => import('@/screens/entity/EntityDetailScreen'), 'EntityDetailScreen');
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
