// ToggleBlock — expandable/collapsible block
import React, { useRef, useEffect, useState } from 'react';
import { View, TextInput, Pressable, Text, StyleSheet } from 'react-native';
import Animated, {
  useSharedValue,
  useAnimatedStyle,
  withTiming,
} from 'react-native-reanimated';
import { useTranslation } from 'react-i18next';
import { colors, fontFamily, fontSize, spacing } from '@/theme';
import type { BlockProps } from '../types';

export const ToggleBlock = React.memo(function ToggleBlock({
  block,
  isActive,
  readOnly,
  onChange,
  onFocus,
  onBackspace,
  onSlashTrigger,
}: BlockProps) {
  const { t } = useTranslation();
  const inputRef = useRef<TextInput>(null);
  const [expanded, setExpanded] = useState(false);
  const rotation = useSharedValue(0);

  useEffect(() => {
    if (isActive && inputRef.current) {
      inputRef.current.focus();
    }
  }, [isActive]);

  const handleToggle = () => {
    const next = !expanded;
    setExpanded(next);
    rotation.value = withTiming(next ? 90 : 0, { duration: 200 });
  };

  const chevronStyle = useAnimatedStyle(() => ({
    transform: [{ rotate: `${rotation.value}deg` }],
  }));

  const handleChangeText = (text: string) => {
    if (text === '/') {
      onSlashTrigger();
      return;
    }
    onChange({ content: text });
  };

  const handleKeyPress = (e: { nativeEvent: { key: string } }) => {
    if (e.nativeEvent.key === 'Backspace' && block.content === '') {
      onBackspace();
    }
  };

  return (
    <View style={styles.container}>
      <View style={styles.header}>
        <Pressable onPress={handleToggle} style={styles.toggleBtn}>
          <Animated.View style={chevronStyle}>
            <Text style={styles.chevron}>▶</Text>
          </Animated.View>
        </Pressable>
        <TextInput
          ref={inputRef}
          style={styles.input}
          value={block.content}
          onChangeText={handleChangeText}
          onFocus={onFocus}
          onKeyPress={handleKeyPress}
          multiline
          editable={!readOnly}
          placeholder={t('editor.toggleHeader')}
          placeholderTextColor={colors.textMuted}
          blurOnSubmit={false}
        />
      </View>
      {expanded && (
        <View style={styles.body}>
          <Text style={styles.bodyText}>
            {/* Toggle children will be rendered by nested blocks in future */}
            Konten toggle kosong
          </Text>
        </View>
      )}
    </View>
  );
});

const styles = StyleSheet.create({
  container: {
    paddingVertical: spacing.xs,
    paddingHorizontal: spacing.lg,
  },
  header: {
    flexDirection: 'row',
    alignItems: 'flex-start',
  },
  toggleBtn: {
    padding: 4,
    marginRight: spacing.xs,
    marginTop: 2,
  },
  chevron: {
    fontSize: 10,
    color: colors.textMuted,
  },
  input: {
    flex: 1,
    fontFamily: fontFamily.documentMedium,
    fontSize: fontSize.md,
    lineHeight: fontSize.md * 1.5,
    color: colors.textPrimary,
    minHeight: 24,
    padding: 0,
  },
  body: {
    marginLeft: 24,
    paddingTop: spacing.xs,
    paddingBottom: spacing.xs,
  },
  bodyText: {
    fontFamily: fontFamily.document,
    fontSize: fontSize.sm,
    color: colors.textMuted,
    fontStyle: 'italic',
  },
});
