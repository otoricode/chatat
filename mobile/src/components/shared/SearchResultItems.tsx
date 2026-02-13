// SearchResultItems — individual search result row components
import React from 'react';
import { View, Text, Pressable, StyleSheet } from 'react-native';
import { Avatar } from '@/components/ui/Avatar';
import { HighlightedText } from '@/components/shared/HighlightedText';
import { formatChatListTime } from '@/lib/timeFormat';
import { colors, fontFamily, fontSize, spacing } from '@/theme';
import type {
  MessageSearchResult,
  DocumentSearchResult,
  ContactSearchResult,
  EntitySearchResult,
} from '@/services/api/search';

// Message result row
type MessageResultProps = {
  item: MessageSearchResult;
  onPress: (item: MessageSearchResult) => void;
};

export function MessageResultItem({ item, onPress }: MessageResultProps) {
  return (
    <Pressable style={styles.row} onPress={() => onPress(item)}>
      <Avatar emoji="💬" size="sm" />
      <View style={styles.content}>
        <View style={styles.topRow}>
          <Text style={styles.title} numberOfLines={1}>
            {item.chatName}
          </Text>
          <Text style={styles.time}>{formatChatListTime(item.createdAt)}</Text>
        </View>
        <Text style={styles.subtitle} numberOfLines={1}>
          {item.senderName}
        </Text>
        <HighlightedText text={item.highlight} />
      </View>
    </Pressable>
  );
}

// Document result row
type DocumentResultProps = {
  item: DocumentSearchResult;
  onPress: (item: DocumentSearchResult) => void;
};

export function DocumentResultItem({ item, onPress }: DocumentResultProps) {
  return (
    <Pressable style={styles.row} onPress={() => onPress(item)}>
      <Avatar emoji={item.icon || '📄'} size="sm" />
      <View style={styles.content}>
        <View style={styles.topRow}>
          <Text style={styles.title} numberOfLines={1}>
            {item.title}
          </Text>
          <Text style={styles.time}>{formatChatListTime(item.updatedAt)}</Text>
        </View>
        {item.locked && (
          <Text style={styles.lockedBadge}>🔒 Terkunci</Text>
        )}
        <HighlightedText text={item.highlight} />
      </View>
    </Pressable>
  );
}

// Contact result row
type ContactResultProps = {
  item: ContactSearchResult;
  onPress: (item: ContactSearchResult) => void;
};

export function ContactResultItem({ item, onPress }: ContactResultProps) {
  return (
    <Pressable style={styles.row} onPress={() => onPress(item)}>
      <Avatar emoji={item.avatar || '👤'} size="sm" />
      <View style={styles.content}>
        <Text style={styles.title} numberOfLines={1}>
          {item.name}
        </Text>
        <Text style={styles.subtitle} numberOfLines={1}>
          {item.phone}
        </Text>
        {item.status ? (
          <Text style={styles.status} numberOfLines={1}>
            {item.status}
          </Text>
        ) : null}
      </View>
    </Pressable>
  );
}

// Entity result row
type EntityResultProps = {
  item: EntitySearchResult;
  onPress: (item: EntitySearchResult) => void;
};

export function EntityResultItem({ item, onPress }: EntityResultProps) {
  return (
    <Pressable style={styles.row} onPress={() => onPress(item)}>
      <Avatar emoji="🏷️" size="sm" />
      <View style={styles.content}>
        <Text style={styles.title} numberOfLines={1}>
          {item.name}
        </Text>
        <Text style={styles.subtitle} numberOfLines={1}>
          {item.type}
        </Text>
      </View>
    </Pressable>
  );
}

// Section header for "Semua" tab
type SectionHeaderProps = {
  title: string;
  count: number;
  onSeeAll?: () => void;
};

export function SectionHeader({ title, count, onSeeAll }: SectionHeaderProps) {
  if (count === 0) return null;
  return (
    <View style={styles.sectionHeader}>
      <Text style={styles.sectionTitle}>{title}</Text>
      {onSeeAll && count > 3 && (
        <Pressable onPress={onSeeAll}>
          <Text style={styles.seeAll}>Lihat Semua</Text>
        </Pressable>
      )}
    </View>
  );
}

const styles = StyleSheet.create({
  row: {
    flexDirection: 'row',
    alignItems: 'flex-start',
    paddingHorizontal: spacing.lg,
    paddingVertical: spacing.md,
    gap: spacing.md,
  },
  content: {
    flex: 1,
  },
  topRow: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    marginBottom: 2,
  },
  title: {
    flex: 1,
    fontFamily: fontFamily.uiSemiBold,
    fontSize: fontSize.md,
    color: colors.textPrimary,
    marginRight: spacing.sm,
  },
  subtitle: {
    fontFamily: fontFamily.ui,
    fontSize: fontSize.sm,
    color: colors.textMuted,
    marginBottom: 2,
  },
  status: {
    fontFamily: fontFamily.ui,
    fontSize: fontSize.xs,
    color: colors.textMuted,
  },
  time: {
    fontFamily: fontFamily.ui,
    fontSize: fontSize.xs,
    color: colors.textMuted,
  },
  lockedBadge: {
    fontFamily: fontFamily.ui,
    fontSize: fontSize.xs,
    color: colors.yellow,
    marginBottom: 2,
  },
  sectionHeader: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    paddingHorizontal: spacing.lg,
    paddingTop: spacing.lg,
    paddingBottom: spacing.sm,
  },
  sectionTitle: {
    fontFamily: fontFamily.uiSemiBold,
    fontSize: fontSize.sm,
    color: colors.textMuted,
    textTransform: 'uppercase',
    letterSpacing: 0.5,
  },
  seeAll: {
    fontFamily: fontFamily.uiMedium,
    fontSize: fontSize.sm,
    color: colors.green,
  },
});
