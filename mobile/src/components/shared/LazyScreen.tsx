// LazyScreen — wrapper for lazily-loaded screen components
import React, { Suspense } from 'react';
import { View, ActivityIndicator, StyleSheet } from 'react-native';
import { colors } from '@/theme';

function LoadingFallback() {
  return (
    <View style={styles.container}>
      <ActivityIndicator size="large" color={colors.green} />
    </View>
  );
}

/**
 * Creates a lazy-loaded screen component using React.lazy + Suspense.
 * Usage:
 *   const LazyDocumentEditor = createLazyScreen(
 *     () => import('@/screens/document/DocumentEditorScreen'), 'DocumentEditorScreen'
 *   );
 */
export function createLazyScreen(
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  factory: () => Promise<Record<string, React.ComponentType<any>>>,
  exportName: string,
) {
  const LazyComponent = React.lazy(async () => {
    const mod = await factory();
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    return { default: mod[exportName] as React.ComponentType<any> };
  });

  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  return function LazyScreen(props: any) {
    return (
      <Suspense fallback={<LoadingFallback />}>
        <LazyComponent {...props} />
      </Suspense>
    );
  };
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    justifyContent: 'center',
    alignItems: 'center',
    backgroundColor: colors.background,
  },
});
