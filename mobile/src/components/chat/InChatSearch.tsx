// InChatSearch — search bar overlay for finding messages within a chat
import React, { useState, useEffect, useRef, useCallback } from 'react';
import {
  View,
  TextInput,
  Text,
  Pressable,
  StyleSheet,
  Animated,
  FlatList,
} from 'react-native';
import { HighlightedText } from '@/components/shared/HighlightedText';
import { searchApi } from '@/services/api/search';
import { useTranslation } from 'react-i18next';
import type { MessageSearchResult } from '@/services/api/search';
import { colors, fontFamily, fontSize, spacing } from '@/theme';

type InChatSearchProps = {
  chatId: string;
  visible: boolean;
  onClose: () => void;
  onResultPress: (messageId: string) => void;
};

export function InChatSearch({
  chatId,
  visible,
  onClose,
  onResultPress,
}: InChatSearchProps) {
  const [query, setQuery] = useState('');
  const [results, setResults] = useState<MessageSearchResult[]>([]);
  const [currentIndex, setCurrentIndex] = useState(0);
  const [isLoading, setIsLoading] = useState(false);
  const { t } = useTranslation();
  const slideAnim = useRef(new Animated.Value(-60)).current;
  const timerRef = useRef<ReturnType<typeof setTimeout> | null>(null);
  const inputRef = useRef<TextInput>(null);

  useEffect(() => {
    Animated.timing(slideAnim, {
      toValue: visible ? 0 : -60,
      duration: 200,
      useNativeDriver: true,
    }).start();

    if (visible) {
      setTimeout(() => inputRef.current?.focus(), 100);
    } else {
      setQuery('');
      setResults([]);
      setCurrentIndex(0);
    }
  }, [visible, slideAnim]);

  const doSearch = useCallback(
    async (q: string) => {
      if (q.length < 2) {
        setResults([]);
        setCurrentIndex(0);
        return;
      }

      setIsLoading(true);
      try {
        const res = await searchApi.searchInChat(chatId, q, 0, 50);
        const data = res.data.data ?? [];
        setResults(data);
        setCurrentIndex(0);
        if (data.length > 0 && data[0]) {
          onResultPress(data[0].id);
        }
      } catch {
        setResults([]);
      } finally {
        setIsLoading(false);
      }
    },
    [chatId, onResultPress],
  );

  useEffect(() => {
    if (timerRef.current) {
      clearTimeout(timerRef.current);
    }
    timerRef.current = setTimeout(() => {
      doSearch(query);
    }, 300);

    return () => {
      if (timerRef.current) {
        clearTimeout(timerRef.current);
      }
    };
  }, [query, doSearch]);

  const goUp = () => {
    if (results.length === 0) return;
    const next = (currentIndex - 1 + results.length) % results.length;
    setCurrentIndex(next);
    const item = results[next];
    if (item) onResultPress(item.id);
  };

  const goDown = () => {
    if (results.length === 0) return;
    const next = (currentIndex + 1) % results.length;
    setCurrentIndex(next);
    const item = results[next];
    if (item) onResultPress(item.id);
  };

  if (!visible) return null;

  return (
    <Animated.View
      style={[styles.container, { transform: [{ translateY: slideAnim }] }]}
    >
      <View style={styles.searchRow}>
        <Text style={styles.icon}>🔍</Text>
        <TextInput
          ref={inputRef}
          style={styles.input}
          value={query}
          onChangeText={setQuery}
          placeholder={t('search.searchInChat')}
          placeholderTextColor={colors.textMuted}
          returnKeyType="search"
        />
        {results.length > 0 && (
          <Text style={styles.counter}>
            {currentIndex + 1}/{results.length}
          </Text>
        )}
        <Pressable onPress={goUp} style={styles.navBtn}>
          <Text style={styles.navIcon}>▲</Text>
        </Pressable>
        <Pressable onPress={goDown} style={styles.navBtn}>
          <Text style={styles.navIcon}>▼</Text>
        </Pressable>
        <Pressable onPress={onClose} style={styles.closeBtn}>
          <Text style={styles.closeIcon}>✕</Text>
        </Pressable>
      </View>
      {results.length > 0 && (
        <FlatList
          data={results}
          keyExtractor={(item) => item.id}
          renderItem={({ item, index }) => (
            <Pressable
              style={[
                styles.resultRow,
                index === currentIndex && styles.resultActive,
              ]}
              onPress={() => {
                setCurrentIndex(index);
                onResultPress(item.id);
              }}
            >
              <Text style={styles.senderName}>{item.senderName}</Text>
              <HighlightedText text={item.highlight} />
            </Pressable>
          )}
          style={styles.resultList}
          keyboardShouldPersistTaps="handled"
        />
      )}
      {query.length >= 2 && results.length === 0 && !isLoading && (
        <View style={styles.noResult}>
          <Text style={styles.noResultText}>
            {t('search.noMessagesFound', { query })}
          </Text>
        </View>
      )}
    </Animated.View>
  );
}

const styles = StyleSheet.create({
  container: {
    position: 'absolute',
    top: 0,
    left: 0,
    right: 0,
    backgroundColor: colors.surface,
    borderBottomWidth: 1,
    borderBottomColor: colors.border,
    zIndex: 100,
    maxHeight: 300,
  },
  searchRow: {
    flexDirection: 'row',
    alignItems: 'center',
    paddingHorizontal: spacing.md,
    paddingVertical: spacing.sm,
    gap: spacing.xs,
  },
  icon: {
    fontSize: 14,
  },
  input: {
    flex: 1,
    fontFamily: fontFamily.ui,
    fontSize: fontSize.md,
    color: colors.textPrimary,
    paddingVertical: spacing.xs,
  },
  counter: {
    fontFamily: fontFamily.ui,
    fontSize: fontSize.xs,
    color: colors.textMuted,
    marginHorizontal: spacing.xs,
  },
  navBtn: {
    padding: spacing.xs,
  },
  navIcon: {
    fontSize: 12,
    color: colors.textMuted,
  },
  closeBtn: {
    padding: spacing.xs,
  },
  closeIcon: {
    fontSize: 16,
    color: colors.textMuted,
    fontWeight: '600',
  },
  resultList: {
    maxHeight: 200,
  },
  resultRow: {
    paddingHorizontal: spacing.lg,
    paddingVertical: spacing.sm,
    borderTopWidth: StyleSheet.hairlineWidth,
    borderTopColor: colors.border,
  },
  resultActive: {
    backgroundColor: colors.surface2,
  },
  senderName: {
    fontFamily: fontFamily.uiMedium,
    fontSize: fontSize.xs,
    color: colors.green,
    marginBottom: 2,
  },
  noResult: {
    paddingHorizontal: spacing.lg,
    paddingVertical: spacing.md,
  },
  noResultText: {
    fontFamily: fontFamily.ui,
    fontSize: fontSize.sm,
    color: colors.textMuted,
    textAlign: 'center',
  },
});
