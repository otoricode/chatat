// Block component props shared across all block types
import type { Block, BlockType } from '@/types/chat';

export interface BlockProps {
  block: Block;
  isActive: boolean;
  readOnly: boolean;
  onChange: (changes: Partial<Block>) => void;
  onFocus: () => void;
  onSubmit: () => void;
  onBackspace: () => void;
  onSlashTrigger: () => void;
}

export interface SlashOption {
  type: BlockType;
  icon: string;
  labelKey: string;
  descriptionKey: string;
}

export const BLOCK_OPTIONS: SlashOption[] = [
  { type: 'paragraph', icon: '📝', labelKey: 'editor.paragraph', descriptionKey: 'editor.paragraphDesc' },
  { type: 'heading1', icon: '𝗛₁', labelKey: 'editor.heading1', descriptionKey: 'editor.heading1Desc' },
  { type: 'heading2', icon: '𝗛₂', labelKey: 'editor.heading2', descriptionKey: 'editor.heading2Desc' },
  { type: 'heading3', icon: '𝗛₃', labelKey: 'editor.heading3', descriptionKey: 'editor.heading3Desc' },
  { type: 'bullet-list', icon: '•', labelKey: 'editor.bulletList', descriptionKey: 'editor.bulletListDesc' },
  { type: 'numbered-list', icon: '1.', labelKey: 'editor.numberedList', descriptionKey: 'editor.numberedListDesc' },
  { type: 'checklist', icon: '☑', labelKey: 'editor.checklist', descriptionKey: 'editor.checklistDesc' },
  { type: 'table', icon: '▦', labelKey: 'editor.table', descriptionKey: 'editor.tableDesc' },
  { type: 'callout', icon: '💡', labelKey: 'editor.callout', descriptionKey: 'editor.calloutDesc' },
  { type: 'code', icon: '⌨', labelKey: 'editor.code', descriptionKey: 'editor.codeDesc' },
  { type: 'toggle', icon: '▶', labelKey: 'editor.toggle', descriptionKey: 'editor.toggleDesc' },
  { type: 'divider', icon: '—', labelKey: 'editor.divider', descriptionKey: 'editor.dividerDesc' },
  { type: 'quote', icon: '❝', labelKey: 'editor.quote', descriptionKey: 'editor.quoteDesc' },
];
