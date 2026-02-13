// BlockEditor — main editor component with FlatList rendering
import React, { useCallback, useMemo, useState, useRef } from 'react';
import {
  View,
  FlatList,
  Pressable,
  StyleSheet,
  KeyboardAvoidingView,
  Platform,
} from 'react-native';
import { colors, spacing } from '@/theme';
import { useEditorStore } from '@/stores/editorStore';
import type { Block, BlockType } from '@/types/chat';
import type { BlockProps } from './types';
import {
  ParagraphBlock,
  HeadingBlock,
  BulletListBlock,
  NumberedListBlock,
  ChecklistBlock,
  QuoteBlock,
  DividerBlock,
  CodeBlock,
  CalloutBlock,
  ToggleBlock,
  TableBlock,
} from './blocks';
import { SlashMenu } from './SlashMenu';
import { FloatingToolbar } from './FloatingToolbar';
import { BlockActionMenu } from './BlockActionMenu';

interface BlockEditorProps {
  readOnly?: boolean;
}

export function BlockEditor({ readOnly = false }: BlockEditorProps) {
  const blocks = useEditorStore((s) => s.blocks);
  const activeBlockId = useEditorStore((s) => s.activeBlockId);
  const showSlashMenu = useEditorStore((s) => s.showSlashMenu);
  const slashFilter = useEditorStore((s) => s.slashFilter);
  const slashBlockId = useEditorStore((s) => s.slashBlockId);
  const showToolbar = useEditorStore((s) => s.showToolbar);

  const setActiveBlock = useEditorStore((s) => s.setActiveBlock);
  const updateBlock = useEditorStore((s) => s.updateBlock);
  const addBlock = useEditorStore((s) => s.addBlock);
  const deleteBlock = useEditorStore((s) => s.deleteBlock);
  const duplicateBlock = useEditorStore((s) => s.duplicateBlock);
  const moveBlock = useEditorStore((s) => s.moveBlock);
  const changeBlockType = useEditorStore((s) => s.changeBlockType);
  const openSlashMenu = useEditorStore((s) => s.openSlashMenu);
  const closeSlashMenu = useEditorStore((s) => s.closeSlashMenu);
  const closeToolbar = useEditorStore((s) => s.closeToolbar);
  const applyFormat = useEditorStore((s) => s.applyFormat);

  const [actionMenuBlockId, setActionMenuBlockId] = useState<string | null>(null);

  const flatListRef = useRef<FlatList>(null);

  // Compute numbered list indices
  const numberedIndices = useMemo(() => {
    const indices: Record<string, number> = {};
    let counter = 0;
    for (const block of blocks) {
      if (block.type === 'numbered-list') {
        counter++;
        indices[block.id] = counter;
      } else {
        counter = 0;
      }
    }
    return indices;
  }, [blocks]);

  const handleSlashSelect = useCallback(
    (type: BlockType) => {
      if (slashBlockId) {
        // If the slash-triggering block is empty, change its type
        const block = blocks.find((b) => b.id === slashBlockId);
        if (block && block.content === '') {
          changeBlockType(slashBlockId, type);
        } else {
          // Add new block after current
          addBlock(type, slashBlockId);
        }
      }
      closeSlashMenu();
    },
    [slashBlockId, blocks, changeBlockType, addBlock, closeSlashMenu],
  );

  const handleBlockAction = useCallback(
    (actionId: string) => {
      if (!actionMenuBlockId) return;

      switch (actionId) {
        case 'duplicate':
          duplicateBlock(actionMenuBlockId);
          break;
        case 'delete':
          deleteBlock(actionMenuBlockId);
          break;
        case 'moveUp':
          moveBlock(actionMenuBlockId, 'up');
          break;
        case 'moveDown':
          moveBlock(actionMenuBlockId, 'down');
          break;
        case 'changeType':
          setActionMenuBlockId(null);
          openSlashMenu(actionMenuBlockId);
          return;
      }
      setActionMenuBlockId(null);
    },
    [actionMenuBlockId, duplicateBlock, deleteBlock, moveBlock, openSlashMenu],
  );

  const renderBlock = useCallback(
    ({ item }: { item: Block }) => {
      const isActive = item.id === activeBlockId;

      const commonProps: BlockProps = {
        block: item,
        isActive,
        readOnly,
        onChange: (changes) => updateBlock(item.id, changes),
        onFocus: () => setActiveBlock(item.id),
        onSubmit: () => {
          // Create same-type block on Enter for list types
          const listTypes = ['bullet-list', 'numbered-list', 'checklist'];
          const nextType = listTypes.includes(item.type) ? item.type : 'paragraph';
          addBlock(nextType as BlockType, item.id);
        },
        onBackspace: () => deleteBlock(item.id),
        onSlashTrigger: () => openSlashMenu(item.id),
      };

      const blockContent = (() => {
        switch (item.type) {
          case 'paragraph':
            return <ParagraphBlock {...commonProps} />;
          case 'heading1':
          case 'heading2':
          case 'heading3':
            return <HeadingBlock {...commonProps} />;
          case 'bullet-list':
            return <BulletListBlock {...commonProps} />;
          case 'numbered-list':
            return (
              <NumberedListBlock {...commonProps} index={numberedIndices[item.id] ?? 1} />
            );
          case 'checklist':
            return <ChecklistBlock {...commonProps} />;
          case 'quote':
            return <QuoteBlock {...commonProps} />;
          case 'divider':
            return <DividerBlock />;
          case 'code':
            return <CodeBlock {...commonProps} />;
          case 'callout':
            return <CalloutBlock {...commonProps} />;
          case 'toggle':
            return <ToggleBlock {...commonProps} />;
          case 'table':
            return <TableBlock {...commonProps} />;
          default:
            return <ParagraphBlock {...commonProps} />;
        }
      })();

      return (
        <Pressable
          style={[styles.blockWrapper, isActive && styles.activeBlock]}
          onLongPress={() => {
            if (!readOnly) setActionMenuBlockId(item.id);
          }}
          delayLongPress={500}
        >
          {isActive && !readOnly && (
            <View style={styles.dragHandle}>
              <View style={styles.dot} />
              <View style={styles.dot} />
              <View style={styles.dot} />
              <View style={styles.dot} />
              <View style={styles.dot} />
              <View style={styles.dot} />
            </View>
          )}
          <View style={styles.blockContent}>{blockContent}</View>
        </Pressable>
      );
    },
    [
      activeBlockId,
      readOnly,
      updateBlock,
      setActiveBlock,
      addBlock,
      deleteBlock,
      openSlashMenu,
      numberedIndices,
    ],
  );

  return (
    <KeyboardAvoidingView
      style={styles.container}
      behavior={Platform.OS === 'ios' ? 'padding' : undefined}
      keyboardVerticalOffset={100}
    >
      <FlatList
        ref={flatListRef}
        data={blocks}
        renderItem={renderBlock}
        keyExtractor={(item) => item.id}
        keyboardShouldPersistTaps="handled"
        keyboardDismissMode="interactive"
        contentContainerStyle={styles.contentContainer}
        removeClippedSubviews={false}
      />

      <SlashMenu
        visible={showSlashMenu}
        filter={slashFilter}
        onSelect={handleSlashSelect}
        onDismiss={closeSlashMenu}
      />

      <FloatingToolbar
        visible={showToolbar}
        onAction={applyFormat}
        onDismiss={closeToolbar}
      />

      <BlockActionMenu
        visible={!!actionMenuBlockId}
        onAction={handleBlockAction}
        onDismiss={() => setActionMenuBlockId(null)}
      />
    </KeyboardAvoidingView>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: colors.background,
  },
  contentContainer: {
    paddingVertical: spacing.sm,
    paddingBottom: 100,
  },
  blockWrapper: {
    flexDirection: 'row',
    alignItems: 'flex-start',
    minHeight: 28,
  },
  activeBlock: {
    backgroundColor: 'rgba(110,231,183,0.04)',
  },
  dragHandle: {
    width: 20,
    paddingTop: spacing.sm,
    paddingLeft: spacing.xs,
    flexDirection: 'row',
    flexWrap: 'wrap',
    gap: 2,
    alignContent: 'center',
    justifyContent: 'center',
  },
  dot: {
    width: 3,
    height: 3,
    borderRadius: 1.5,
    backgroundColor: colors.textMuted,
  },
  blockContent: {
    flex: 1,
  },
});
